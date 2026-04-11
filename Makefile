.PHONY: build build-cli build-ctl test lint clean install docs snapshot release

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS  = -s -w \
  -X github.com/itunified-io/dbx/internal/version.Version=$(VERSION) \
  -X github.com/itunified-io/dbx/internal/version.Commit=$(COMMIT) \
  -X github.com/itunified-io/dbx/internal/version.Date=$(DATE)

build: build-cli build-ctl

build-cli:
	go build -ldflags "$(LDFLAGS)" -o bin/dbxcli ./cmd/dbxcli

build-ctl:
	go build -ldflags "$(LDFLAGS)" -o bin/dbxctl ./cmd/dbxctl

test:
	go test -race -cover ./...

lint:
	golangci-lint run ./...

docs:
	@rm -rf docs/cli/
	go run ./cmd/docgen

snapshot:
	goreleaser release --snapshot --clean

release:
	goreleaser release --clean

clean:
	rm -rf bin/ dist/

install: build
	cp bin/dbxcli $(GOPATH)/bin/dbxcli
	cp bin/dbxctl $(GOPATH)/bin/dbxctl
