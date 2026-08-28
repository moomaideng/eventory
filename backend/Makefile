.PHONY: dev run migrate reset seed test build clean tidy

# Main Application Commands
dev:
	air

run:
	go run cmd/api/main.go

migrate:
	go run cmd/migrate/main.go

reset:
	go run cmd/migrate/main.go -reset

seed:
	go run cmd/seed/main.go

# Go Utilities
test:
	go test -v ./...

build:
	go build -o bin/api cmd/api/main.go

clean:
	rm -rf bin/ tmp/

tidy:
	go mod tidy
	go fmt ./...