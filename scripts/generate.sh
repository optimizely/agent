#!/usr/bin/env bash
# Script to generate API clients and documentation via the provided swagger spec.

# This script currently depends on OpenAPI, https://swagger.io/specification/ and [OpenAPI Generator](https://github.com/openapitools/openapi-generator)
# (a [fork](https://github.com/OpenAPITools/openapi-generator/blob/master/docs/migration-from-swagger-codegen.md) of swagger-codegen).

# - To install the OpenAPI Generator on OSX via Homebrew (e.g. `brew install openapi-generator`)
# - To determine which generators are available you can execute `openapi-generator` without any arguments or refer to the generator source [docs](https://github.com/OpenAPITools/openapi-generator/blob/master/docs/generators/README.md):

# usage: scripts/generate.sh <GENERATOR_NAME>

set -u
# Generator name which can be sourced via `openapi-generator`
NAME=$1

INPUT_FILE="$PWD/api/openapi-spec/openapi.yaml"

mkdir -p generated/$NAME
cd generated/$NAME
openapi-generator generate -g $NAME -i $INPUT_FILE
cd -

Echo "Generated API at: $INPUT_FILE"
