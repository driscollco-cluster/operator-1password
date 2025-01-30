package main

import (
	"github.com/driscollco-cluster/operator-1password/internal/conf"
	"github.com/driscollco-cluster/operator-1password/internal/crds"
	"github.com/driscollco-cluster/operator-1password/internal/operator"
	operatorLib "github.com/driscollco-core/kubernetes-operator"
	logLib "github.com/driscollco-core/log"
	"github.com/driscollco-core/service"
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
		log.SetLogger(logLib.NewLogr(s.Log().Child("operator", "person")))
		op := operatorLib.New("person-controller", operator.Reconcile(s.Log().Child("operator", "person")))
		if err := op.Start("crds.driscollco", "v1", &crds.Secret{}, &crds.SecretList{}); err != nil {
			s.Log().Error("unable to start the operator", "error", err.Error())
			os.Exit(0)
		}
	}()
	s.Run()
}
