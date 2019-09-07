[![Build Status](https://travis-ci.com/optimizely/sidedoor.svg?token=y3xM1z7bQsqHX2NTEhps&branch=master)](https://travis-ci.com/optimizely/sidedoor)
[![codecov](https://codecov.io/gh/optimizely/sidedoor/branch/master/graph/badge.svg?token=UabuO3fxyA)](https://codecov.io/gh/optimizely/sidedoor)
# Optimizely Sidedoor
Exploratory project for developing a service version of the Optimizely SDK.

## Package Structure
Following best practice for go project layout as defined [here](https://github.com/golang-standards/project-layout)

* **api** - OpenAPI/Swagger specs, JSON schema files, protocol definition files.
* **cmd** - Main applications for this project.
* **bin** - Compiled application binaries.
* **pkg** - Library code that can be used by other applications
* **scripts** - Scripts to perform various build, install, analysis, etc operations.

## Make targets
The following `make` targets can be used to build and run the application:
* **build** - builds sidedoor and installs binary in bin/sidedoor
* **clean** - runs `go clean` and removes the bin/ dir
* **install** - runs `go get` to install all dependencies
* **generate-api** - generates APIs from the swagger spec
* **lint** - runs `golangci-lint` linters defined in `.golangci.yml` file
* **run** - builds and executes the sidedoor binary
* **test** - recursively tests all .go files

## Running locally
Currently the Optimizely SDK Key is sourced from an `SDK_KEY` environment variable. For local development you can export your `SDK_KEY` or prefix the `make run` command.

Ex:
```
SDK_KEY=<YOUR-KEY-KEY> make run
```

This file will get loaded via the `Makefile` configuration script.

## Client Generation

### Prerequisites

Install go on OSX:
```
brew install go
```

This repo currently depends heavily on [OpenAPI](https://swagger.io/specification/) and [OpenAPI Generator](https://github.com/openapitools/openapi-generator) (a [fork](https://github.com/OpenAPITools/openapi-generator/blob/master/docs/migration-from-swagger-codegen.md) of swagger-codegen).

To install the OpenAPI Generator on OSX:
```
brew install openapi-generator
```

To determine which generators are available you can execute `openapi-generator` without any arguments or refer to the generator source [docs](https://github.com/OpenAPITools/openapi-generator/blob/master/docs/generators/README.md):

Types of generators are either CLIENT, SERVER, DOCUMENTATION, SCHEMA and CONFIG.

### Generating
You can use the helper script `generate.sh` to experiment with the various generated assets.
```
scripts/generate.sh <GENERATOR_NAME>
```
We also provide a Make task `generate-api`:
```
make generate-api ARG=<GENERATOR_NAME>
```
