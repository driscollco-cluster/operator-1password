package handlerBasic

import (
	"fmt"
	"github.com/driscollco-cluster/operator-1password/internal/conf"
	"github.com/driscollco-core/http-router"
)

const (
	InfoGotArgument = "got url argument"
	InfoNoArgument  = "got request without url argument"
)

func Handle(request router.Request) router.Response {
	log := request.Log().Child("handler", "basic")
	log.Info("gcp project", "project", conf.Config.GCP.ProjectId)
	if request.ArgExists("one") {
		log.Info(InfoGotArgument, "one", request.GetArg("one"))
		return request.Success(fmt.Sprintf("received: %s", request.GetArg("one")))
	}
	log.Info(InfoNoArgument)
	return request.Success("OK")
}
