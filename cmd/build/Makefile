repoName = $(shell yq '.image.artifactRegistry.repo' helm/values.yaml)
serviceName = $(shell yq '.service.name' helm/values.yaml)
project = $(shell yq '.image.artifactRegistry.project' helm/values.yaml)
region = $(shell yq '.image.artifactRegistry.hostname' helm/values.yaml)
tag = $(shell yq '.image.tag' helm/values.yaml)
namespace = $(shell yq '.service.namespace' helm/values.yaml)

.PHONY: build
build:
	@env GOOS=linux CGO_ENABLED=0 go build -o service ../main.go

.PHONY: docker
docker:
	@docker build -t $(serviceName):$(tag) .
	@docker tag $(serviceName):$(tag) $(region)/$(project)/$(repoName)/$(serviceName):$(tag)
	@docker push $(region)/$(project)/$(repoName)/$(serviceName):$(tag)
	@rm -f service

.PHONY: deploy
deploy:
	@helm upgrade --history-max 1 --install $(serviceName) helm --namespace $(namespace) --create-namespace