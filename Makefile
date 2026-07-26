# Pinned so a future `go` directive never silently pulls a newer toolchain.
export GOTOOLCHAIN=local

.PHONY: build vet test run tidy

build:
	go build ./...

vet:
	go vet ./...

test:
	go test ./...

run:
	go run ./cmd/server

tidy:
	go mod tidy
