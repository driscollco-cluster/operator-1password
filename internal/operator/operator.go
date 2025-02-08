package operator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	onepassword "github.com/driscollco-cluster/1password"
	"github.com/driscollco-cluster/operator-1password/internal/conf"
	"github.com/driscollco-cluster/operator-1password/internal/crds"
	"github.com/driscollco-core/log"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

const (
	finalizer       = "opsecrets.crds.driscoll.co"
	errorToSuppress = "resourceVersion should not be set on objects to be created"
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
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	theLog := o.log.Child(
		"secret.source", fmt.Sprintf("%s/%s/%s", opsecret.Spec.Source.Vault, opsecret.Spec.Source.Item, opsecret.Spec.Source.Section),
		"opsecret.location", fmt.Sprintf("%s/%s", req.Namespace, req.Name))

	if !opsecret.ObjectMeta.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(opsecret, finalizer) {
			theLog.Info("opsecret is deleting")
			for _, namespace := range opsecret.Spec.Secret.Namespaces {
				childSecret := &corev1.Secret{}
				secretKey := types.NamespacedName{
					Name:      opsecret.Spec.Secret.Name,
					Namespace: namespace,
				}
				if err := k8sClient.Get(ctx, secretKey, childSecret); err != nil {
					theLog.Error("unable to fetch child secret", "error", err.Error(),
						"secret.location", fmt.Sprintf("%s/%s", namespace, opsecret.Spec.Secret.Name))
					continue
				}
				if err := k8sClient.Delete(ctx, childSecret); err != nil {
					theLog.Error("error deleting child secret", "error", err.Error(),
						"secret.location", fmt.Sprintf("%s/%s", namespace, opsecret.Spec.Secret.Name))
					continue
				}
			}
			controllerutil.RemoveFinalizer(opsecret, finalizer)
			if err := k8sClient.Update(ctx, opsecret); err != nil {
				theLog.Error("error removing finalizer for opsecret", "error", err.Error())
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(opsecret, finalizer) {
		controllerutil.AddFinalizer(opsecret, finalizer)
		if err := k8sClient.Update(ctx, opsecret); err != nil {
			theLog.Error("error setting finalizer for opsecret", "error", err.Error())
			return ctrl.Result{}, err
		}
		return o.getRequeue(opsecret), nil
	}

	if opsecret.Status.Events == nil {
		opsecret.Status.Events = []crds.Event{}
	}

	item, err := o.client.GetItem(opsecret.Spec.Source.Vault, opsecret.Spec.Source.Item)
	if err != nil {
		theLog.Info("error fetching item from 1Password", "error", err.Error())
		return ctrl.Result{}, err
	}

	section, ok := item.Content[opsecret.Spec.Source.Section]
	if !ok {
		theLog.Info("section was not found when looking for updates to secret")
		return o.getRequeue(opsecret), nil
	}
	if !o.updateRequired(opsecret, section) {
		return o.getRequeue(opsecret), nil
	}

	k8sSecret := &corev1.Secret{}
	switch opsecret.Spec.Secret.SecretType {
	case "docker":
		k8sSecret, err = o.getDockerSecret(opsecret, section)
		if err != nil {
			return ctrl.Result{}, err
		}
	default:
		k8sSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: opsecret.Spec.Secret.Name,
			},
			Type:       corev1.SecretTypeOpaque,
			StringData: make(map[string]string),
		}

		for _, key := range opsecret.Spec.Secret.Keys {
			found, ok := section.Values[key.From]
			if !ok {
				foundAsFile, ok := section.Files[key.From]
				if !ok {
					o.log.Error("could not find matching key in section", "key", key.From)
					return ctrl.Result{}, nil
				}
				fileContent, err := o.client.FileContent(foundAsFile)
				if err != nil {
					o.log.Error("error trying to retrieve contents of file", "file", foundAsFile.Name, "error", err.Error())
					return ctrl.Result{}, nil
				}
				k8sSecret.StringData[key.To] = string(fileContent)
			} else {
				k8sSecret.StringData[key.To] = found.Value
			}
		}
	}

	// Check if the opsecret already exists
	for _, namespace := range opsecret.Spec.Secret.Namespaces {
		k8sSecret.Namespace = namespace
		existingSecret := &corev1.Secret{}
		err = k8sClient.Get(ctx, types.NamespacedName{Name: k8sSecret.Name, Namespace: namespace}, existingSecret)
		if err != nil && apierrors.IsNotFound(err) {
			err = k8sClient.Create(ctx, k8sSecret)
			if err != nil {
				if err.Error() != errorToSuppress {
					theLog.Error("Failed to create secret", "error", err.Error(),
						"secret.location", fmt.Sprintf("%s/%s", namespace, opsecret.Spec.Secret.Name))
				}
				return ctrl.Result{}, err
			}
			theLog.Info("created new secret", "secret.location", fmt.Sprintf("%s/%s", namespace, opsecret.Spec.Secret.Name))

			opsecret.Status.Events = append(opsecret.Status.Events, crds.Event{
				Timestamp:   metav1.Now(),
				OpTimestamp: metav1.NewTime(section.LastUpdated),
				Type:        "create",
				Message:     "Secret created from 1Password data",
			})
		} else if err == nil {
			if reflect.DeepEqual(existingSecret.StringData, k8sSecret.StringData) {
				continue
			}
			existingSecret.StringData = k8sSecret.StringData
			err = k8sClient.Update(ctx, existingSecret)
			if err != nil {
				theLog.Error("failed to update secret", "error", err.Error(),
					"secret.location", fmt.Sprintf("%s/%s", namespace, opsecret.Spec.Secret.Name))
				return ctrl.Result{}, err
			}
			theLog.Info("updated secret", "secret.location", fmt.Sprintf("%s/%s", namespace, opsecret.Spec.Secret.Name))

			opsecret.Status.Events = append(opsecret.Status.Events, crds.Event{
				Timestamp:   metav1.Now(),
				OpTimestamp: metav1.NewTime(section.LastUpdated),
				Type:        "update",
				Message:     "secret has been updated to reflect changes in 1Password",
			})

			// Find pods that reference this opsecret
			podList := &corev1.PodList{}
			err = k8sClient.List(ctx, podList, client.InNamespace(req.Namespace))
			if err != nil {
				theLog.Error("failed to list pods", "error", err.Error(),
					"secret.location", fmt.Sprintf("%s/%s", namespace, opsecret.Spec.Secret.Name))
				return ctrl.Result{}, err
			}

			// Iterate over pods and delete those using the opsecret
			for _, pod := range podList.Items {
				if isPodUsingSecret(&pod, opsecret.Spec.Secret.Name) {
					err = k8sClient.Delete(ctx, &pod)
					if err != nil {
						theLog.Error("failed to delete pod", "pod", pod.Name, "error", err.Error(),
							"secret.location", fmt.Sprintf("%s/%s", namespace, opsecret.Spec.Secret.Name))
					}
				}
			}

		} else {
			theLog.Error("error checking for existing secret", "error", err.Error(),
				"secret.location", fmt.Sprintf("%s/%s", namespace, opsecret.Spec.Secret.Name))
			return ctrl.Result{}, err
		}
	}

	if opsecret.Status.Events == nil {
		opsecret.Status.Events = []crds.Event{}
	}

	opsecret.Status.LastUpdated = metav1.NewTime(time.Now())
	err = k8sClient.Status().Update(ctx, opsecret)
	if err != nil {
		theLog.Error("failed to update the last updated time for opsecret", "error", err.Error())
		return ctrl.Result{}, err
	}

	return o.getRequeue(opsecret), nil
}

func (o operator) updateRequired(opsecret *crds.OpSecret, section onepassword.Section) bool {
	if opsecret.Status.LastUpdated.Time.After(section.LastUpdated) {
		return false
	}

	if opsecret.Spec.Secret.SecretType == "docker" {
		return true
	}

	for _, keyMapping := range opsecret.Spec.Secret.Keys {
		if opsecret.Status.LastUpdated.Time.After(section.Values[keyMapping.From].LastUpdated) {
			return false
		}
	}
	return true
}

func (o operator) getRequeue(opsecret *crds.OpSecret) ctrl.Result {
	if opsecret.Spec.Secret.RefreshSeconds >= conf.Config.Secrets.Refresh.MinIntervalSeconds {
		return ctrl.Result{RequeueAfter: time.Second * time.Duration(opsecret.Spec.Secret.RefreshSeconds)}
	}
	return ctrl.Result{}
}

func (o operator) getDockerSecret(opsecret *crds.OpSecret, section onepassword.Section) (*corev1.Secret, error) {
	if opsecret.Spec.Secret.SecretType != "docker" {
		return nil, errors.New("wrong secret type: " + opsecret.Spec.Secret.SecretType)
	}

	if len(opsecret.Spec.Secret.Keys) < 1 {
		return nil, errors.New("no keys defined")
	}

	file, ok := section.Files[opsecret.Spec.Secret.Keys[0].From]
	if !ok {
		return nil, fmt.Errorf("specified soure file does not exist in section: %s", opsecret.Spec.Secret.Keys[0])
	}

	tokenData, err := o.client.FileContent(file)
	if err != nil {
		return nil, errors.New("missing file content: token.json")
	}

	dockerConfig := map[string]interface{}{
		"auths": map[string]map[string]string{
			opsecret.Spec.Secret.Keys[0].To: {
				"username": "_json_key",
				"password": string(tokenData),
				"email":    "test@jdd.email",
			},
		},
	}
	dockerConfigJson, err := json.Marshal(dockerConfig)
	if err != nil {
		return nil, err
	}

	// Create the Secret
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: opsecret.Spec.Secret.Name,
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			".dockerconfigjson": dockerConfigJson,
		},
	}, nil
}

func isPodUsingSecret(pod *corev1.Pod, secretName string) bool {
	for _, pullSecret := range pod.Spec.ImagePullSecrets {
		if pullSecret.Name == secretName {
			return true
		}
	}

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
