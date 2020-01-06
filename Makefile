# The name of the executable (default is current directory name)
TARGET := "optimizely"
APP_VERSION ?= $(shell git describe --tags)
.DEFAULT_GOAL := help

COVER_FILE := cover.out

# Go parameters
GO111MODULE:=on
GOCMD=go
GOBIN=bin
GOPATH=$(shell $(GOCMD) env GOPATH)
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test -race
GOGET=$(GOCMD) get
GOLINT=golangci-lint
BINARY_UNIX=$(TARGET)_unix

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

# Use linker flags to strip debugging info from the binary.
# -s Omit the symbol table and debug information.
# -w Omit the DWARF symbol table.
LDFLAGS=-ldflags "-s -w -X main.Version=${APP_VERSION}"

all: test build ## all
$(TARGET):
	GO111MODULE=$(GO111MODULE) $(GOBUILD) $(LDFLAGS) -o $(GOBIN)/$(TARGET) cmd/$(TARGET)/main.go

build: $(TARGET) ## builds and installs binary in bin/
	@true

cover: ## runs test suite with coverage profiling
	GO111MODULE=$(GO111MODULE) $(GOTEST) ./... -coverprofile=$(COVER_FILE)

cover-html: cover ## generates test coverage html report
	$(GOCMD) tool cover -html=$(COVER_FILE)

clean: ## runs `go clean` and removes the bin/ dir
	GO111MODULE=$(GO111MODULE) $(GOCLEAN) --modcache
	rm -rf $(GOBIN)

install: ## installs all dev and ci dependencies
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b $(GOPATH)/bin v1.19.0

lint: ## runs `golangci-lint` linters defined in `.golangci.yml` file
	$(GOLINT) run --out-format=tab --tests=false ./...

run: $(TARGET) ## builds and executes the TARGET binary
	$(GOBIN)/$(TARGET)

test: ## recursively tests all .go files
	GO111MODULE=$(GO111MODULE) $(GOTEST) ./...

include scripts/Makefile.ci

help: ## help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
