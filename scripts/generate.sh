#!/usr/bin/env bash
set -u
# Generator name which can be sourced via `openapi-generator`
NAME=$1

CURRENT_DIR=$PWD
INPUT_FILE="$PWD/api/openapi-spec/openapi.yaml"

mkdir $NAME
cd $NAME
openapi-generator generate -g $NAME -i $INPUT_FILE
cd CURRENT_DIR

Echo "Generated API at: $INPUT_FILE"
