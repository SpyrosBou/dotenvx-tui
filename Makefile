BINARY=dotenvx-tui
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

.PHONY: build build-all install lint test clean

build:
	go build -ldflags "-s -w" -o $(BINARY) .

build-all:
	GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o dist/$(BINARY)-darwin-arm64 .
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o dist/$(BINARY)-darwin-amd64 .
	GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o dist/$(BINARY)-linux-arm64 .
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o dist/$(BINARY)-linux-amd64 .

install: build
	mv $(BINARY) $(GOBIN)/$(BINARY)

lint:
	golangci-lint run ./...

test:
	go test ./...

clean:
	rm -f $(BINARY)
	rm -rf dist/
