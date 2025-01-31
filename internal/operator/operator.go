package operator

import (
	"context"
	onepassword "github.com/driscollco-cluster/1password"
	"github.com/driscollco-cluster/operator-1password/internal/conf"
	"github.com/driscollco-cluster/operator-1password/internal/crds"
	"github.com/driscollco-core/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type Operator interface {
	Reconcile(ctx context.Context, req ctrl.Request, k8sClient client.Client, recorder record.EventRecorder) (ctrl.Result, error)
}

func New(log log.Log) Operator {
	return operator{
		client: onepassword.NewClient(conf.Config.OnePassword.Api.Url, conf.Config.OnePassword.Api.Token),
		log:    log,
	}
}

type operator struct {
	client onepassword.Client
	log    log.Log
}

func (o operator) Reconcile(ctx context.Context, req ctrl.Request, k8sClient client.Client, recorder record.EventRecorder) (ctrl.Result, error) {
	o.log.Info("reconcile started for resource", "name", req.Name, "namespace", req.Namespace)

	secret := &crds.ExternalSecret{}
	if err := k8sClient.Get(ctx, req.NamespacedName, secret); err != nil {
		o.log.Error("unable to fetch secret resource", "error", err.Error())
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	o.log.Info("request to fetch secret from 1Password", "vault", secret.Spec.Source.Vault,
		"item", secret.Spec.Source.Item, "section", secret.Spec.Source.Section)

	k8sSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secret.Spec.Secret.Name,
			Namespace: secret.Spec.Secret.Namespace,
		},
		Type:       corev1.SecretTypeOpaque,
		StringData: make(map[string]string),
	}

	for _, keyMapping := range secret.Spec.Secret.Keys {
		info, err := o.client.GetKey(secret.Spec.Source.Vault, secret.Spec.Source.Item, secret.Spec.Source.Section, keyMapping.From)
		if err != nil {
			o.log.Error("error fetching information from 1Password", "error", err.Error())
			return ctrl.Result{}, nil
		}
		k8sSecret.StringData[keyMapping.To] = info.Value
	}

	// Check if the secret already exists
	existingSecret := &corev1.Secret{}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: k8sSecret.Name, Namespace: k8sSecret.Namespace}, existingSecret)

	if err != nil && errors.IsNotFound(err) {
		// Secret does not exist, create it
		err = k8sClient.Create(ctx, k8sSecret)
		if err != nil {
			o.log.Error("Failed to create Kubernetes Secret", "error", err.Error())
			return ctrl.Result{}, err
		}
		o.log.Info("Successfully created Kubernetes Secret", "name", k8sSecret.Name, "namespace", k8sSecret.Namespace)
	} else if err == nil {
		// Secret exists, update it
		existingSecret.StringData = k8sSecret.StringData
		err = k8sClient.Update(ctx, existingSecret)
		if err != nil {
			o.log.Error("Failed to update Kubernetes Secret", "error", err.Error())
			return ctrl.Result{}, err
		}
		o.log.Info("Successfully updated Kubernetes Secret", "name", k8sSecret.Name, "namespace", k8sSecret.Namespace)

		// Find pods that reference this secret
		podList := &corev1.PodList{}
		err = k8sClient.List(ctx, podList, client.InNamespace(req.Namespace))
		if err != nil {
			o.log.Error("failed to list pods", "error", err.Error())
			return ctrl.Result{}, err
		}

		// Iterate over pods and delete those using the secret
		for _, pod := range podList.Items {
			if isPodUsingSecret(&pod, secret.Name) {
				o.log.Info("Deleting pod using updated secret", "pod", pod.Name)
				err = k8sClient.Delete(ctx, &pod)
				if err != nil {
					o.log.Error("failed to delete pod", "pod", pod.Name, "error", err.Error())
				}
			}
		}

	} else {
		o.log.Error("Error checking for existing Kubernetes Secret", "error", err.Error())
		return ctrl.Result{}, err
	}

	secret.Status.LastUpdated = metav1.NewTime(time.Now())
	err = k8sClient.Status().Update(ctx, secret)
	if err != nil {
		o.log.Error("failed to update the last updated time for secret", "error", err.Error())
		return ctrl.Result{}, err
	}

	if secret.Spec.Secret.RefreshSeconds >= conf.Config.Secrets.Refresh.MinIntervalSeconds {
		return ctrl.Result{RequeueAfter: time.Second * time.Duration(secret.Spec.Secret.RefreshSeconds)}, nil
	}
	return ctrl.Result{}, nil
}

func isPodUsingSecret(pod *corev1.Pod, secretName string) bool {
	for _, container := range pod.Spec.Containers {
		// Check envFrom
		for _, envFrom := range container.EnvFrom {
			if envFrom.SecretRef != nil && envFrom.SecretRef.Name == secretName {
				return true
			}
		}

		// Check individual env variables
		for _, envVar := range container.Env {
			if envVar.ValueFrom != nil && envVar.ValueFrom.SecretKeyRef != nil && envVar.ValueFrom.SecretKeyRef.Name == secretName {
				return true
			}
		}
	}

	// Check if the secret is mounted as a volume
	for _, volume := range pod.Spec.Volumes {
		if volume.Secret != nil && volume.Secret.SecretName == secretName {
			return true
		}
	}

	return false
}
