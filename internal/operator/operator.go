package operator

import (
	"context"
	"github.com/driscollco-cluster/operator-1password/internal/crds"
	operatorLib "github.com/driscollco-core/kubernetes-operator"
	"github.com/driscollco-core/log"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Reconcile(log log.Log) operatorLib.ReconcileFunc {
	return func(ctx context.Context, req ctrl.Request, k8sClient client.Client, recorder record.EventRecorder) (ctrl.Result, error) {
		log.Info("reconcile started for resource", "name", req.Name, "namespace", req.Namespace)

		secret := &crds.Secret{}
		if err := k8sClient.Get(ctx, req.NamespacedName, secret); err != nil {
			log.Error("unable to fetch secret resource", "error", err.Error())
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}

		log.Info("request to fetch secret from 1Password", "vault", secret.Spec.SourceVault,
			"item", secret.Spec.SourceItem, "section", secret.Spec.SourceSection, "key", secret.Spec.SourceKey)
		return ctrl.Result{}, nil
	}
}
