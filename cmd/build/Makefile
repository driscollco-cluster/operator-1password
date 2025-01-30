repoName = $(shell yq '.image.artifactRegistry.repo' config.yml)
serviceName = $(shell yq '.service.name' config.yml)
project = $(shell yq '.image.artifactRegistry.project' config.yml)
region = $(shell yq '.image.artifactRegistry.hostname' config.yml)
tag = $(shell yq '.image.tag' config.yml)
namespace = $(shell yq '.service.namespace' config.yml)

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