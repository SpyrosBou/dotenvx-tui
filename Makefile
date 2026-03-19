BINARY=dotenvx-tui
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GO_LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build build-all install lint test clean release release-dry-run

build:
	go build $(GO_LDFLAGS) -o $(BINARY) .

build-all:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(GO_LDFLAGS) -o dist/$(BINARY)-darwin-arm64 .
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(GO_LDFLAGS) -o dist/$(BINARY)-darwin-amd64 .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(GO_LDFLAGS) -o dist/$(BINARY)-linux-arm64 .
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(GO_LDFLAGS) -o dist/$(BINARY)-linux-amd64 .

install: build
	mv $(BINARY) $(GOBIN)/$(BINARY)

lint:
	golangci-lint run ./...

test:
	go test ./...

clean:
	rm -f $(BINARY)
	rm -rf dist/

release-dry-run:
	goreleaser release --snapshot --clean

release:
	goreleaser release --clean
