[![Build Status](https://travis-ci.com/optimizely/sidedoor.svg?token=y3xM1z7bQsqHX2NTEhps&branch=master)](https://travis-ci.com/optimizely/sidedoor)
[![codecov](https://codecov.io/gh/optimizely/sidedoor/branch/master/graph/badge.svg?token=UabuO3fxyA)](https://codecov.io/gh/optimizely/sidedoor)
# Optimizely Sidedoor
Sidedoor is the Optimizely Full Stack Service which exposes the functionality of a Full Stack SDK as
a highly available and distributed application.

## Package Structure
Following best practice for go project layout as defined [here](https://github.com/golang-standards/project-layout)

* **api** - OpenAPI/Swagger specs, JSON schema files, protocol definition files.
* **bin** - Compiled application binaries.
* **cmd** - Main applications for this project.
* **docs** - User documentation files.
* **pkg** - Library code that can be used by other applications.
* **scripts** - Scripts to perform various build, install, analysis, etc operations.

## Make targets
The following `make` targets can be used to build and run the application:
* **build** - builds sidedoor and installs binary in bin/sidedoor
* **clean** - runs `go clean` and removes the bin/ dir
* **install** - installs all dev and ci dependencies
* **generate-api** - generates APIs from the swagger spec
* **lint** - runs `golangci-lint` linters defined in `.golangci.yml` file
* **run** - builds and executes the sidedoor binary
* **test** - recursively tests all .go files

## Prerequisites
Install go on OSX:
```
brew install go
```

## Client Generation
This repo currently depends on [OpenAPI](https://swagger.io/specification/) and [OpenAPI Generator](https://github.com/openapitools/openapi-generator) (a [fork](https://github.com/OpenAPITools/openapi-generator/blob/master/docs/migration-from-swagger-codegen.md) of swagger-codegen).

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

## Credits

This software is distributed with code from the following open source projects:

gohistogram
Copyright (c) 2013 VividCortex
LICENSE (MIT):https://github.com/VividCortex/gohistogram/blob/master/LICENSE

chi
Copyright (c) 2015-present Peter Kieltyka
LICENSE (MIT):https://github.com/go-chi/chi/blob/master/LICENSE

chi-render
Copyright (c) 2016-Present
LICENSE (MIT):https://github.com/go-chi/render/blob/master/LICENSE

go-kit
Copyright (c) 2015 Peter Bourgon
LICENSE (MIT):https://github.com/go-kit/kit/blob/master/LICENSE

uuid
Copyright (c) 2009,2014 Google Inc. All rights reserved.
License (BSD-3): https://github.com/nsqio/nsq/blob/master/LICENSE

nsq
License (MIT): https://github.com/nsqio/nsq/blob/master/LICENSE

Optimizely go-sdk
License (Apache-2): https://github.com/optimizely/go-sdk/blob/master/LICENSE

concurrent-map
Copyright (c) 2014 streamrail
License (MIT): https://github.com/orcaman/concurrent-map/blob/master/LICENSE

zerolog
Copyright (c) 2017 Olivier Poitrey
License (MIT): https://github.com/rs/zerolog/blob/master/LICENSE

nsq-go
Copyright (c) 2016 Segment
LICENSE (MIT):https://github.com/segmentio/nsq-go/blob/master/LICENSE

viper
Copyright (c) 2014 Steve Francia
LICENSE (MIT):https://github.com/spf13/viper/blob/master/LICENSE

testify
Copyright (c) 2012-2018 Mat Ryer and Tyler Bunnell.
License (MIT): https://github.com/stretchr/testify/blob/master/LICENSE

net
Copyright (c) 2009 The Go Authors. All rights reserved.
https://github.com/golang/sync/blob/master/LICENSE

sync
Copyright (c) 2009 The Go Authors. All rights reserved.
https://github.com/golang/sync/blob/master/LICENSE

sys
Copyright (c) 2009 The Go Authors. All rights reserved.
https://github.com/golang/sync/blob/master/LICENSE
