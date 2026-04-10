.PHONY: build build-cli build-ctl test lint clean install

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS_CLI := -ldflags "-X main.version=$(VERSION)"
LDFLAGS_CTL := -ldflags "-X github.com/itunified-io/dbx/internal/version.Version=$(VERSION)"

build: build-cli build-ctl

build-cli:
	go build $(LDFLAGS_CLI) -o bin/dbxcli ./cmd/dbxcli

build-ctl:
	go build $(LDFLAGS_CTL) -o bin/dbxctl ./cmd/dbxctl

test:
	go test -race -cover ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/ dist/

install: build
	cp bin/dbxcli $(GOPATH)/bin/dbxcli
	cp bin/dbxctl $(GOPATH)/bin/dbxctl
