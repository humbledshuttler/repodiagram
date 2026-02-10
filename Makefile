BINARY_NAME=repodiagram
VERSION?=0.1.0
BUILD_DIR=build

.PHONY: build build-all clean install test

build:
	go build -ldflags="-s -w" -o $(BINARY_NAME) .

install:
	go install .

test:
	go test -v ./...

clean:
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)

build-all: clean
	mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

deps:
	go mod tidy

fmt:
	go fmt ./...

lint:
	golangci-lint run
