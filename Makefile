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
GOLINT=$(GOPATH)/bin/golangci-lint
BINARY_UNIX=$(TARGET)_unix

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

# Use linker flags to strip debugging info from the binary.
# -s Omit the symbol table and debug information.
# -w Omit the DWARF symbol table.
LDFLAGS=-ldflags "-s -w -X main.Version=${APP_VERSION} -X github.com/optimizely/go-sdk/v2/pkg/event.ClientName=Agent -X github.com/optimizely/go-sdk/v2/pkg/event.Version=${APP_VERSION}"
.PHONY: all lint clean

all: test lint build ## runs the test, lint and build targets

$(TARGET): check-go static
	$(GOBUILD) $(LDFLAGS) -o $(GOBIN)/$(TARGET) cmd/optimizely/main.go

build: $(TARGET) check-go ## builds and installs binary in bin/
	@true

check-go:
ifndef GOPATH
	$(error "go is not available please install golang version 1.24.0+, https://golang.org/dl/")
endif

clean: check-go ## runs `go clean` and removes the bin/ dir
	$(GOCLEAN) --modcache
	rm -rf $(GOBIN)

cover: check-go static ## runs test suite with coverage profiling
#   Run tests with coverage on all packages
	$(GOTEST) ./... -coverprofile=$(COVER_FILE).tmp
#   Exclude test helpers, utility files, and generated code from coverage metrics:
#   - optimizelytest/ files are test helpers, not production code
#   - redis.go pubsub implementation is difficult to test in CI
#   - generate_secret is a utility command not part of core functionality
#   - statik.go is generated code that shouldn't affect coverage metrics
	grep -v -E "optimizelytest/|pubsub/redis.go|cmd/generate_secret/|statik/statik.go" $(COVER_FILE).tmp > $(COVER_FILE)
	rm $(COVER_FILE).tmp


setup: check-go ## installs all dev and ci dependencies, but does not install golang 
#   Install golangci-lint
	@echo "Installing golangci-lint v1.64.2..."
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v1.64.2
	@echo "Installing statik..."
	go install github.com/rakyll/statik@latest

lint: check-go static ## runs `golangci-lint` linters defined in `.golangci.yml` file
	$(GOLINT) run --out-format=tab --tests=false ./...

run: $(TARGET) ## builds and executes the TARGET binary
	$(GOBIN)/$(TARGET)

stop:	## stops TARGET binary process
	pkill -f "$(GOBIN)/$(TARGET)"

static: check-go
	$(GOPATH)/bin/statik -src=web/static -f

test: check-go static ## recursively tests all .go files
	$(GOTEST) ./...

include scripts/Makefile.ci

# Generate secret helper
GEN_SECRET_TARGET := "generate_secret"

$(GEN_SECRET_TARGET): check-go
	$(GOBUILD) $(LDFLAGS) -o $(GOBIN)/$(GEN_SECRET_TARGET) cmd/generate_secret/main.go

build_generate_secret: $(GEN_SECRET_TARGET) ## builds the GEN_SECRET_TARGET binary
	@true

generate_secret: $(GEN_SECRET_TARGET) ## builds and executes the GEN_SECRET_TARGET binary
	$(GOBIN)/$(GEN_SECRET_TARGET)

help: ## help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

test-acceptance:
	export OPTIMIZELY_SERVER_BATCHREQUESTS_OPERATIONSLIMIT='3' && \
	export OPTIMIZELY_CLIENT_USERPROFILESERVICE='{"default":"in-memory","services":{"in-memory":{"storagestrategy":"fifo"}}}' && \
	export OPTIMIZELY_CLIENT_ODP_SEGMENTSCACHE='{"default":"redis","services":{"redis":{"host":"localhost:6379","password":"","timeout":"0s","database": 0}}}' && \
	make clean && \
	make setup && \
	make run & \
	bash scripts/wait_for_agent_to_start.sh && \
	pytest -vv -rA --diff-symbols tests/acceptance/test_acceptance/ \
	-k "not test_decide__feature_no_ups and not test_decide__flag_key_parameter_no_ups" --host "$(MYHOST)" && \
	make stop
