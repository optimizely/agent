# The name of the executable (default is current directory name)
TARGET := $(shell basename "$(PWD)")
.DEFAULT_GOAL := help

GO111MODULE:=on

# Go parameters
GOCMD=go
GOBIN=bin
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOLINT=golangci-lint
BINARY_UNIX=$(TARGET)_unix

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

# Use linker flags to strip debugging info from the binary.
# -s Omit the symbol table and debug information.
# -w Omit the DWARF symbol table.
LDFLAGS=-ldflags "-s -w"

all: test build ## all
$(TARGET):
	GO111MODULE=$(GO111MODULE) $(GOBUILD) $(LDFLAGS) -o $(GOBIN)/$(TARGET) cmd/$(TARGET)/main.go

build: $(TARGET) ## builds and installs binary in bin/
	@true

cover:
	GO111MODULE=$(GO111MODULE) $(GOTEST) -v -race ./... -coverprofile=cover.out

cover-html: cover
	$(GOCMD) tool cover -html=cover.out

clean: ## runs `go clean` and removes the bin/ dir
	$(GOCLEAN)
	rm -rf $(GOBIN)

generate-api: ## generates APIs from the swagger spec
	scripts/generate.sh $(ARG)

install: ## runs `go get` to install all dependencies
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint

lint: install ## runs `golangci-lint` linters defined in `.golangci.yml` file
	$(GOLINT) run --out-format=tab --tests=false pkg/...
	$(GOLINT) run --out-format=tab --tests=false cmd/...

run: $(TARGET) ## builds and executes the TARGET binary
	$(GOBIN)/$(TARGET)

test: ## recursively tests all .go files
	GO111MODULE=$(GO111MODULE) $(GOTEST) -v ./...

include scripts/Makefile.ci

help: ## help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
