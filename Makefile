include .env

# The name of the executable (default is current directory name)
TARGET := $(shell basename "$(PWD)")
.DEFAULT_GOAL := $(TARGET)  # TODO make a help

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

# Use linker flags to provide version/build settings to the target
LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"

all: test build
$(TARGET): 
	$(GOBUILD) $(LDFLAGS) -o $(GOBIN)/$(TARGET) cmd/$(TARGET)/main.go
build: $(TARGET)
	@true
test: 
	$(GOTEST) -v ./...
clean: 
	$(GOCLEAN)
	rm -rf $(GOBIN)
run: $(TARGET)
	$(GOBIN)/$(TARGET)
lint:
	$(GOLINT) run --out-format=tab --tests=false pkg/...

install:
	$(GOGET) github.com/go-chi/chi
	$(GOGET) github.com/go-chi/render
	$(GOGET) github.com/optimizely/go-sdk/optimizely/client
	$(GOGET) github.com/optimizely/go-sdk/optimizely/entities
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint
generate-api:
	scripts/generate.sh $(ARG)
help:
	echo "TODO"
