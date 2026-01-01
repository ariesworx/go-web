go-unit-test:
	go test -v ./...

go-vet:
	go vet ./...

go-lint:
	golangci-lint run

check: go-vet go-lint go-unit-test