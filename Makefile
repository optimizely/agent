include .env

# Go parameters
GOCMD=go
GOBIN=bin
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=sidedoor
BINARY_UNIX=$(BINARY_NAME)_unix

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

all: test build
build: 
	$(GOBUILD) -o $(GOBIN)/$(BINARY_NAME) cmd/sidedoor/main.go 
test: 
	$(GOTEST) -v ./...
clean: 
	$(GOCLEAN)
	rm -rf $(GOBIN)
run: build
	$(GOBIN)/$(BINARY_NAME)
deps:
	$(GOGET) github.com/go-chi/chi
	$(GOGET) github.com/go-chi/render
	$(GOGET) github.com/optimizely/go-sdk/optimizely/client
	$(GOGET) github.com/optimizely/go-sdk/optimizely/entities
generate-api:
	scripts/generate.sh $(ARG)

