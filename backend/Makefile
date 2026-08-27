.PHONY: build fmt run test vet

build:
	go build -o bin/api ./cmd/api

fmt:
	gofmt -w $$(find . -name '*.go' -type f)

run:
	go run ./cmd/api

test:
	go test ./...

vet:
	go vet ./...
