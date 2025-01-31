package operator

import (
	"context"
	onepassword "github.com/driscollco-cluster/1password"
	"github.com/driscollco-cluster/operator-1password/internal/conf"
	"github.com/driscollco-cluster/operator-1password/internal/crds"
	"github.com/driscollco-core/log"
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

	o.log.Info("request to fetch secret from 1Password", "vault", secret.Spec.SourceVault,
		"item", secret.Spec.SourceItem, "section", secret.Spec.SourceSection, "key", secret.Spec.SourceKey)

	info, err := o.client.GetItem(secret.Spec.SourceVault, secret.Spec.SourceItem)
	if err != nil {
		o.log.Error("error fetching secret from 1Password", "error", err.Error())
		return ctrl.Result{}, nil
	}
	o.log.Info("got information from 1Password", info.Values[secret.Spec.SourceSection][secret.Spec.SourceKey])
	return ctrl.Result{}, nil
}
