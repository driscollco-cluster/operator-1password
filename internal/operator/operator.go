package operator

import (
	"context"
	"fmt"
	onepassword "github.com/driscollco-cluster/1password"
	"github.com/driscollco-cluster/operator-1password/internal/conf"
	"github.com/driscollco-cluster/operator-1password/internal/crds"
	"github.com/driscollco-core/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

type Operator interface {
	Reconcile(ctx context.Context, req ctrl.Request, k8sClient client.Client, recorder record.EventRecorder, scheme *runtime.Scheme) (ctrl.Result, error)
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

func (o operator) Reconcile(ctx context.Context, req ctrl.Request, k8sClient client.Client, recorder record.EventRecorder, scheme *runtime.Scheme) (ctrl.Result, error) {
	opsecret := &crds.OpSecret{}
	if err := k8sClient.Get(ctx, req.NamespacedName, opsecret); err != nil {
		if apierrors.IsNotFound(err) {
			o.log.Info("opsecret has been deleted", "name", req.Name, "namespace", req.Namespace)
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if opsecret.Status.Events == nil {
		opsecret.Status.Events = []crds.Event{}
	}

	k8sSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opsecret.Spec.Secret.Name,
			Namespace: req.Namespace,
		},
		Type:       corev1.SecretTypeOpaque,
		StringData: make(map[string]string),
	}

	if err := controllerutil.SetControllerReference(opsecret, k8sSecret, scheme); err != nil {
		o.log.Error("error setting owner reference for opsecret", "error", err.Error())
		return ctrl.Result{}, err
	}

	latestUpdate := time.Time{}
	for _, keyMapping := range opsecret.Spec.Secret.Keys {
		info, err := o.client.GetKey(opsecret.Spec.Source.Vault, opsecret.Spec.Source.Item, opsecret.Spec.Source.Section, keyMapping.From)
		if err != nil {
			o.log.Error("error fetching information from 1Password", "error", err.Error())
			return ctrl.Result{}, nil
		}
		k8sSecret.StringData[keyMapping.To] = info.Value
		if info.LastUpdated.After(latestUpdate) {
			latestUpdate = info.LastUpdated
		}
	}

	if opsecret.Status.LastUpdated.Time.After(latestUpdate) {
		if opsecret.Spec.Secret.RefreshSeconds >= conf.Config.Secrets.Refresh.MinIntervalSeconds {
			return ctrl.Result{RequeueAfter: time.Second * time.Duration(opsecret.Spec.Secret.RefreshSeconds)}, nil
		}
		return ctrl.Result{}, nil
	}

	// Check if the opsecret already exists
	existingSecret := &corev1.Secret{}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: k8sSecret.Name, Namespace: k8sSecret.Namespace}, existingSecret)
	if err != nil && errors.IsNotFound(err) {
		err = k8sClient.Create(ctx, k8sSecret)
		if err != nil {
			o.log.Error("Failed to create secret", "error", err.Error())
			return ctrl.Result{}, err
		}
		o.log.Info("created new secret", "name", k8sSecret.Name, "namespace", k8sSecret.Namespace,
			"opsecret", opsecret.Name,
			"source", fmt.Sprintf("%s/%s/%s", opsecret.Spec.Source.Vault, opsecret.Spec.Source.Item, opsecret.Spec.Source.Section))

		opsecret.Status.Events = append(opsecret.Status.Events, crds.Event{
			Timestamp:   metav1.Now(),
			OpTimestamp: metav1.NewTime(latestUpdate),
			Type:        "create",
			Message:     "Secret created from 1Password data",
		})
	} else if err == nil {
		// Secret exists, update it
		existingSecret.StringData = k8sSecret.StringData
		err = k8sClient.Update(ctx, existingSecret)
		if err != nil {
			o.log.Error("failed to update secret", "error", err.Error())
			return ctrl.Result{}, err
		}
		o.log.Info("updated secret", "name", k8sSecret.Name, "namespace", k8sSecret.Namespace, "opsecret", opsecret.Name)

		opsecret.Status.Events = append(opsecret.Status.Events, crds.Event{
			Timestamp:   metav1.Now(),
			OpTimestamp: metav1.NewTime(latestUpdate),
			Type:        "update",
			Message:     "secret has been updated to reflect changes in 1Password",
		})

		// Find pods that reference this opsecret
		podList := &corev1.PodList{}
		err = k8sClient.List(ctx, podList, client.InNamespace(req.Namespace))
		if err != nil {
			o.log.Error("failed to list pods", "error", err.Error())
			return ctrl.Result{}, err
		}

		// Iterate over pods and delete those using the opsecret
		for _, pod := range podList.Items {
			if isPodUsingSecret(&pod, opsecret.Name) {
				o.log.Info("Deleting pod using updated opsecret", "pod", pod.Name)
				err = k8sClient.Delete(ctx, &pod)
				if err != nil {
					o.log.Error("failed to delete pod", "pod", pod.Name, "error", err.Error())
				}
			}
		}

	} else {
		o.log.Error("error checking for existing secret", "error", err.Error())
		return ctrl.Result{}, err
	}

	if opsecret.Status.Events == nil {
		opsecret.Status.Events = []crds.Event{}
	}

	opsecret.Status.LastUpdated = metav1.NewTime(time.Now())
	err = k8sClient.Status().Update(ctx, opsecret)
	if err != nil {
		o.log.Error("failed to update the last updated time for opsecret", "error", err.Error())
		return ctrl.Result{}, err
	}

	if opsecret.Spec.Secret.RefreshSeconds >= conf.Config.Secrets.Refresh.MinIntervalSeconds {
		return ctrl.Result{RequeueAfter: time.Second * time.Duration(opsecret.Spec.Secret.RefreshSeconds)}, nil
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
