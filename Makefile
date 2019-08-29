# The name of the executable (default is current directory name)
TARGET := $(shell basename "$(PWD)")
.DEFAULT_GOAL := $(TARGET)  # TODO make a help

GO111MODULE:=on

# Go parameters
GOCMD=go
GOBIN=bin
GOBUILD=GO111MODULE=$(GO111MODULE) $(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=GO111MODULE=$(GO111MODULE) $(GOCMD) test
GOGET=$(GOCMD) get
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
test: build
	$(GOTEST) -v ./...
clean: 
	$(GOCLEAN)
	rm -rf $(GOBIN)
run: $(TARGET)
	$(GOBIN)/$(TARGET)
generate-api:
	scripts/generate.sh $(ARG)
help:
	echo "TODO"
