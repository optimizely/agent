# Optimizely Sidedoor
Exploratory project for developing a service version of the Optimizely SDK.

## Prerequisites
This repo currently depends heavily on [OpenAPI](https://swagger.io/specification/) and [OpenAPI Generator](https://github.com/openapitools/openapi-generator).

To install the OpenAPI Generator on OSX:
```
brew install openapi-generator
```

## Code Generation
To determine which generators are available:
```
openapi-generator
```
Types of generators are either CLIENT, SERVER, DOCUMENTATION, SCHEMA and CONFIG.

You can use the helper script `generate.sh` to experiment with the various generated assets.
```
./generate.sh <GENERATOR_NAME>
```
