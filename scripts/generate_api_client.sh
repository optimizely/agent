#!/usr/bin/env bash

set -e

mkdir -p optimizely_rest_api
cd optimizely_rest_api
rm -f swagger.json*
wget "https://api.optimizely.com/v2/swagger.json"
openapi-generator generate -g go -i $PWD/swagger.json --skip-validate-spec
echo "Generated REST API client"
rm -f swagger.json*
cd -
