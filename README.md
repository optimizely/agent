[![Build Status](https://travis-ci.com/optimizely/sidedoor.svg?token=y3xM1z7bQsqHX2NTEhps&branch=master)](https://travis-ci.com/optimizely/sidedoor)
[![codecov](https://codecov.io/gh/optimizely/sidedoor/branch/master/graph/badge.svg?token=UabuO3fxyA)](https://codecov.io/gh/optimizely/sidedoor)
# Optimizely Agent
Optimizely Agent is the Optimizely Full Stack Service which exposes the functionality of a Full Stack SDK as
a highly available and distributed web application.

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
* **build** - builds optimizely and installs binary in bin/optimizely
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

This software is used with additional code that is separately downloaded by you. These components are subject to their own license terms which you should review carefully.

Gohistogram
(c) 2013 VividCortex 
License (MIT): github.com/VividCortex/gohistogram

Chi
(c) 2015-present Peter Kieltyka (https://github.com/pkieltyka), Google Inc. 
License (MIT): github.com/go-chi/chi

chi-render 
(c) 2016-Present https://github.com/go-chi ‑ authors
License (MIT): github.com/go-chi/render

go-kit 
(c) 2015 Peter Bourgon
License (MIT): github.com/go-kit/kit

guuid
(c) 2009,2014 Google Inc. All rights reserved. 
License (BSD 3-Clause): github.com/google/uuid

nsq 
Matt Reiferson and Jehiah Czebotar
License (MIT): github.com/nsqio/nsq

optimizely go sdk 
(c) 2016-2017, Optimizely, Inc. and contributors
License (Apache 2): github.com/optimizely/go-sdk

concurrent-map 
(c) 2014 streamrail
License (MIT): github.com/orcaman/concurrent-map

zerolog 
(c) 2017 Olivier Poitrey
License (MIT): github.com/rs/zerolog

nsq-go 
(c) 2016 Segment
License (MIT): github.com/segmentio/nsq-go

viper 
(c) 2014 Steve Francia
License (MIT): github.com/spf13/viper

testify
(c) 2012-2018 Mat Ryer and Tyler Bunnell
License (MIT): github.com/stretchr/testify

net 
(c) 2009 The Go Authors
License (BSD 3-Clause): https://github.com/golang/net
 
sync 
(c) 2009 The Go Authors
License (BSD 3-Clause): https://github.com/golang/sync

sys 
(c) 2009 The Go Authors
License (BSD 3-Clause): https://github.com/golang/sys
 
## Apache Copyright Notice
Copyright 2019-present, Optimizely, Inc. and contributors

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS,   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
