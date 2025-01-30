.PHONY: run
run:
	@cd cmd; go run main.go

#
# This should indicate if your code will pass Github Actions automated tests
#
.PHONY: ready
ready: coverage check-formatting lint test govet

.PHONY: update
update:
	@go get -u -t ./...; go mod tidy
	@make mocks

.PHONY: clean
clean:
	@gofmt -w .

.PHONY: lint
lint:
	@cd internal; golangci-lint run
	@cd cmd; golangci-lint run

.PHONY: mocks
mocks:
	@go get go.uber.org/mock/mockgen
	@go get golang.org/x/tools/internal/gocommand
	@go get golang.org/x/tools/internal/imports
	@cd internal/interfaces; go generate
	@go mod tidy

#
# In support of Github Actions automated checks
#

defaultCodeCoverageMinimum = 20
.PHONY: coverage
coverage:
	@cd internal && \
		go test -v -coverpkg=./... -coverprofile=profile.cov ./... > /dev/null 2>&1 && \
		grep -v 'interfaces/' profile.cov | sed '/interfaces/d' > profile2.cov && mv profile2.cov profile.cov && \
		grep -v 'mocks/' profile.cov | sed '/mocks/d' > profile2.cov && mv profile2.cov profile.cov && \
		grep -v 'main.go' profile.cov | sed '/main.go/d' > profile2.cov && mv profile2.cov profile.cov && \
		percentage=$$(go tool cover -func profile.cov | grep -o 'total:.*' | grep -o '[0-9.]\+%' | awk -F'.' '{print $$1}' | tr -d '%') && \
		rm profile.cov && \
        echo "$$percentage" | awk -v expected="$${codeCoverageMinimum:-$(defaultCodeCoverageMinimum)}" '{ if ($$1 >= expected) { print "Code coverage is", $$1 "%", "which matches or exceeds the required level of", expected "%"; } else { print "Code coverage falls below the requirement. Coverage must be at least", expected "%", "but is only", $$1 "%"; exit 1; } }'

.PHONY: check-formatting
check-formatting:
	@cd internal; if ! test -z "$$(gofmt -l .)"; then echo "Code is not formatted. Please run gofmt -w ."; exit 1; fi
	@cd cmd; if ! test -z "$$(gofmt -l .)"; then echo "Code is not formatted. Please run gofmt -w ."; exit 1; fi

.PHONY: test
test:
	@cd internal; go test ./...
	@cd cmd; go test ./...

.PHONY: govet
govet:
	@cd internal; go vet ./...
	@cd cmd; go vet ./...