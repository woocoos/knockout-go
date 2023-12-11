golangci-lint:
	golangci-lint run

IntegrationDIR := $(CURDIR)/integration
integration-gen:
	cd $(IntegrationDIR) && GOWORK=off go generate ./...

integration-test:
	cd $(IntegrationDIR) && go test ./...