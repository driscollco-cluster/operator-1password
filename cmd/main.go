package main

import (
	"github.com/driscollco-cluster/go-service-rest/internal/conf"
	"github.com/driscollco-cluster/go-service-rest/internal/handlers/basic"
	router "github.com/driscollco-core/http-router"
	processGroupInterfaces "github.com/driscollco-core/process-group/interfaces"
	"github.com/driscollco-core/service"
	"os"
)

func main() {
	s := service.New("example service")
	s.Router().Get("/template", handlerBasic.Handle)
	s.Router().Get("/template/:one", handlerBasic.Handle)
	s.Router().WithShutdownFunc(func(resources router.Resources) {
		resources.Log().Info("http router shutodwn function processing")
	})

	if err := s.ProcessGroup().Process(pinger, "pinger"); err != nil {
		s.Log().Error("unable to launch pinger process", "error", err.Error())
		os.Exit(0)
	}
	s.ProcessGroup().WithShutdownFunc(func(resources processGroupInterfaces.Resources) {
		resources.Log().Info("processgroup shutdown function processing")
	})
	if err := s.Config().Populate(&conf.Config); err != nil {
		s.Log().Error("unable to populate config", "error", err.Error())
		os.Exit(0)
	}
	s.Run()
}

func pinger(process processGroupInterfaces.Process) error {
	process.Log().Info("ping")
	return nil
}
