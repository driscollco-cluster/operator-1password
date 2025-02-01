package main

import (
	"github.com/driscollco-cluster/operator-1password/internal/conf"
	"github.com/driscollco-cluster/operator-1password/internal/crds"
	"github.com/driscollco-cluster/operator-1password/internal/operator"
	operatorLib "github.com/driscollco-core/kubernetes-operator"
	"github.com/driscollco-core/service"
	"github.com/go-logr/logr"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func main() {
	s := service.New("1Password Custom Operator")

	if err := s.Config().Populate(&conf.Config); err != nil {
		s.Log().Error("unable to populate config", "error", err.Error())
		os.Exit(0)
	}

	go func() {
		log.SetLogger(logr.Discard())
		actualOp := operator.New(s.Log().Child("operator", "operator-opsecrets"))
		op := operatorLib.New("operator-opsecrets", actualOp.Reconcile)
		if err := op.Start("crds.driscoll.co", "v1", &crds.OpSecret{}, &crds.OpSecretList{}); err != nil {
			s.Log().Error("unable to start the operator", "error", err.Error())
			os.Exit(0)
		}
	}()
	s.Run()
}
