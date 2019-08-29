include .env

# The name of the executable (default is current directory name)
TARGET := $(shell basename "$(PWD)")
.DEFAULT_GOAL := help

# Go parameters
GOCMD=go
GOBIN=bin
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_UNIX=$(TARGET)_unix

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

# Use linker flags to provide version/build settings to the target
LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"

all: test build ## all
$(TARGET):
	$(GOBUILD) $(LDFLAGS) -o $(GOBIN)/$(TARGET) cmd/$(TARGET)/main.go

build: $(TARGET) ## build
	@true

clean: ## clean
	$(GOCLEAN)
	rm -rf $(GOBIN)

devops_build_fpm_centos: ## build fpm_centos image for packaging
	docker build -f scripts/dockerfiles/Dockerfile.fpm_centos -t fpm_centos ${GOPATH}/bin

devops_build_fpm_ubuntu: ## build fpm_centos image for packaging
	docker build -f scripts/dockerfiles/Dockerfile.fpm_ubuntu -t fpm_ubuntu ${GOPATH}/bin

devops_get_fpm_centos: ## get generated rpm
	docker run -v $(PWD):/output -it fpm_centos bash -c "cp *.rpm /output"

devops_get_fpm_ubuntu: ## get generated deb
	docker run -v $(PWD):/output -it fpm_ubuntu bash -c "cp *.deb /output"

generate-api: ## generate-api
	scripts/generate.sh $(ARG)

install: ## install
	$(GOGET) github.com/go-chi/chi
	$(GOGET) github.com/go-chi/render
	$(GOGET) github.com/optimizely/go-sdk/optimizely/client
	$(GOGET) github.com/optimizely/go-sdk/optimizely/entities

run: $(TARGET) ## run
	$(GOBIN)/$(TARGET)

test: ## test
	$(GOTEST) -v ./...

.PHONY: all build test clean run install generate-api help

help:  ## help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
