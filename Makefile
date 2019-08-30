# The name of the executable (default is current directory name)
TARGET := $(shell basename "$(PWD)")
.DEFAULT_GOAL := $(TARGET)  # TODO make a help

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

all: test build
$(TARGET): 
	GO111MODULE=$(GO111MODULE) $(GOBUILD) $(LDFLAGS) -o $(GOBIN)/$(TARGET) cmd/$(TARGET)/main.go
build: $(TARGET)
	@true
test: build
	GO111MODULE=$(GO111MODULE) $(GOTEST) -v ./...
clean: 
	$(GOCLEAN)
	rm -rf $(GOBIN)
run: $(TARGET)
	$(GOBIN)/$(TARGET)
lint:
	$(GOLINT) run --out-format=tab --tests=false pkg/...

install:
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint
generate-api:
	scripts/generate.sh $(ARG)
help:
	echo "TODO"
