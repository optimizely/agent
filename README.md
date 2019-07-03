# sidedoor
Exploratory project for developing a service version of the Optimizely SDK.

# Prerequisits
This repo currently depends heavily on openapi and openapi codegen.

To install the openapi generator on OSX:
```
brew install openapi-generator
```

# Generate
To determine which generators are available:
```
openapi-generator
```
Types of generators are either CLIENT, SERVER, DOCUMENTATION, SCHEMA and CONFIG.

You can use the helper script `generate.sh` to experiment with the various generated assets.
```
./generate.sh <GENERATOR_NAME>
```
