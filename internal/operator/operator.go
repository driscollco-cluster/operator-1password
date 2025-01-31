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
		k8sSecret.StringData[keyMapping.To] = info
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
	} else {
		o.log.Error("Error checking for existing Kubernetes Secret", "error", err.Error())
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
