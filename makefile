golangci-lint:
	golangci-lint run

integration-gen:
	cd ./integration && GOWORK=off go generate ./...

integration-test:
	cd ./integration && go test ./...