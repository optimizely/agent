#!/usr/bin/env bash
set -u
# Generator name which can be sourced via `openapi-generator`
NAME=$1

INPUT_FILE="$PWD/api/openapi-spec/openapi.yaml"

mkdir -p generated/$NAME
cd generated/$NAME
openapi-generator generate -g $NAME -i $INPUT_FILE
cd -

Echo "Generated API at: $INPUT_FILE"
