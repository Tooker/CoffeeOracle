.PHONY: test lint run build tidy

test:
	go test ./...

lint:
	go vet ./...

run:
	go run ./cmd/server

tidy:
	go mod tidy

build:
	go build ./cmd/server
