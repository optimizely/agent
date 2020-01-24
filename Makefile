# The name of the executable (default is current directory name)
TARGET := "optimizely"
APP_VERSION ?= $(shell git describe --tags 2> /dev/null)
.DEFAULT_GOAL := help

COVER_FILE := cover.out

# Go parameters
GO111MODULE:=on
GOCMD=GO111MODULE=$(GO111MODULE) go
GOBIN=bin
GOPATH:=$(shell $(GOCMD) env GOPATH 2> /dev/null)
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
LDFLAGS=-ldflags "-s -w -X main.Version=${APP_VERSION} -X github.com/optimizely/go-sdk/pkg/event.ClientName=Agent -X github.com/optimizely/go-sdk/pkg/event.Version=${APP_VERSION}"
.PHONY: all lint clean

all: test lint build ## runs the test, lint and build targets

$(TARGET): check-go
	$(GOBUILD) $(LDFLAGS) -o $(GOBIN)/$(TARGET) cmd/main.go

build: $(TARGET) check-go ## builds and installs binary in bin/
	@true

check-go:
ifndef GOPATH
	$(error "go is not available please install golang version 1.13+, https://golang.org/dl/")
endif

clean: check-go ## runs `go clean` and removes the bin/ dir
	$(GOCLEAN) --modcache
	rm -rf $(GOBIN)

cover: check-go ## runs test suite with coverage profiling
	$(GOTEST) ./... -coverprofile=$(COVER_FILE)

cover-html: cover ## generates test coverage html report
	$(GOCMD) tool cover -html=$(COVER_FILE)

install: check-go ## installs all dev and ci dependencies, but does not install golang
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b $(GOPATH)/bin v1.19.0

lint: check-go ## runs `golangci-lint` linters defined in `.golangci.yml` file
	$(GOLINT) run --out-format=tab --tests=false ./...

run: $(TARGET) ## builds and executes the TARGET binary
	$(GOBIN)/$(TARGET)

test: check-go ## recursively tests all .go files
	$(GOTEST) ./...

include scripts/Makefile.ci

help: ## help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
