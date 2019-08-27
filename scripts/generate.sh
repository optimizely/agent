#!/usr/bin/env bash
set -u
# Generator name which can be sourced via `openapi-generator`
NAME=$1

mkdir $NAME
cd $NAME
openapi-generator generate -g $NAME -i ../openapi.yaml
