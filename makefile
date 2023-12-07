golangci-lint:
	golangci-lint run

integration-gen:
	cd ./integration && go generate ./...

integration-test:
	cd ./integration && go test ./...