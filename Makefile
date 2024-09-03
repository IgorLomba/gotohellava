.DEFAULT_GOAL := build

.PHONY: fmt vet build

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

build: vet
	@echo "Building for $(OS)..."
ifeq ($(OS),Windows_NT)
	go build -o bin/ava.exe .
else
	go build -o bin/ava .
endif
