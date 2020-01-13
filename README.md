[![Build Status](https://travis-ci.com/optimizely/agent.svg?token=y3xM1z7bQsqHX2NTEhps&branch=master)](https://travis-ci.com/optimizely/agent)
[![codecov](https://codecov.io/gh/optimizely/agent/branch/master/graph/badge.svg?token=UabuO3fxyA)](https://codecov.io/gh/optimizely/agent)
# Optimizely Agent
Optimizely Agent is the Optimizely Full Stack Service which exposes the functionality of a Full Stack SDK as
a highly available and distributed web application.

## Package Structure
Following best practice for go project layout as defined [here](https://github.com/golang-standards/project-layout)

* **api** - OpenAPI/Swagger specs, JSON schema files, protocol definition files.
* **bin** - Compiled application binaries.
* **cmd** - Main applications for this project.
* **config** - Application configuration.
* **docs** - User documentation files.
* **pkg** - Library code that can be used by other applications.
* **scripts** - Scripts to perform various build, install, analysis, etc operations.

## Make targets
The following `make` targets can be used to build and run the application:
* **build** - builds optimizely and installs binary in bin/optimizely
* **clean** - runs `go clean` and removes the bin/ dir
* **cover** - runs test suite with coverage profiling
* **cover-html** - generates test coverage html report
* **install** - installs all dev and ci dependencies, but does not install golang
* **lint** - runs `golangci-lint` linters defined in `.golangci.yml` file
* **run** - builds and executes the optimizely binary
* **test** - recursively tests all .go files

## Prerequisites
Optimizely Agent is implemented in [Golang](https://golang.org/). Golang is required for developing and compiling from source.
Installers and binary archives for most platforms can be downloaded directly from the Go [downloads](https://golang.org/dl/) page.

## Running Optimizely from source
Once Go is installed, the Optimizely Agent can be started via the following `make` command:
```bash
make run
```
This will start the Optimizely Agent with the default configuration in the foreground.

## Configuration Options
Optimizely Agent configuration can be overridden by a yaml configuration file provided at runtime.

By default the configuration file will be sourced from the current active directory `e.g. ./config.yaml`.
Alternative configuration locations can be specified at runtime via environment variable or command line flag.
```bash
OPTIMIZELY_CONFIG_FILENAME=config.yaml make run
```
The default configuration can be found [here](config.yaml).

Below is a comprehensive list of available configuration properties.

|Property Name|Env Variable|Description|
|---|---|---|
|config.filename|OPTIMIZELY_CONFIG_FILENAME|Location of the configuration YAML file. Default: ./config.yaml|
|log.level|OPTIMIZELY_LOG_LEVEL|The log [level](https://github.com/rs/zerolog#leveled-logging) for the agent. Default: info|
|log.pretty|OPTIMIZELY_LOG_PRETTY|Flag used to set colorized console output as opposed to structured json logs. Default: false|
|server.readtimeout|OPTIMIZELY_SERVER_READTIMEOUT|The maximum duration for reading the entire body. Default: “5s”|
|server.writetimeout|OPTIMIZELY_SERVER_WRITETIMEOUT|The maximum duration before timing out writes of the response. Default: “10s”|
|admin.author|OPTIMIZELY_ADMIN_AUTHOR|Agent version. Default: Optimizely Inc.|
|admin.name|OPTIMIZELY_ADMIN_NAME|Agent name. Default: optimizely|
|admin.port|OPTIMIZELY_ADMIN_PORT|Admin listener port. Default: 8088|
|admin.version|OPTIMIZELY_ADMIN_VERSION|Agent version. Default: `git describe --tags`|
|api.port|OPTIMIZELY_API_PORT|Api listener port. Default: 8080|
|api.maxconns|OPTIMIZLEY_API_MAXCONNS|Maximum number of concurrent requests|
|webhook.port|OPTIMIZELY_WEBHOOK_PORT|Webhook listener port: Default: 8085|
|webhook.projects.<*projectId*>.sdkKeys|N/A|Comma delimited list of SDK Keys applicable to the respective projectId|
|webhook.projects.<*projectId*>.secret|N/A|Webhook secret used to validate webhook requests originating from the respective projectId|
|webhook.projects.<*projectId*>.skipSignatureCheck|N/A|Boolean to indicate whether the signature should be validated. TODO remove in favor of empty secret.|

## Metrics

The Metrics API exposes telemetry data of the running Optimizely Agent. The core runtime metrics are exposed via the go expvar package. Documentation for the various statistics can be found as part of the [mstats](https://golang.org/src/runtime/mstats.go) package.

Example Request:
```bash
curl localhost:8088/metrics
```
Example Response:
```
{
	"cmdline": [
		"bin/optimizely"
	],
	"memstats": {
		"Alloc": 924136,
		"TotalAlloc": 924136,
		"Sys": 71893240,
		"Lookups": 0,
		"Mallocs": 4726,
		"Frees": 172,
		"HeapAlloc": 924136,
		...
	},
	...
}
```
Custom metrics are also provided for the individual service endpoints and follow the pattern of:

```properties
"timers.<metric-name>.counts": 0,
"timers.<metric-name>.responseTime": 0,
"timers.<metric-name>.responseTimeHist.p50": 0,
"timers.<metric-name>.responseTimeHist.p90": 0,
"timers.<metric-name>.responseTimeHist.p95": 0,
"timers.<metric-name>.responseTimeHist.p99": 0,
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

optimizely go sdk 
(c) 2016-2017, Optimizely, Inc. and contributors
License (Apache 2): github.com/optimizely/go-sdk

concurrent-map 
(c) 2014 streamrail
License (MIT): github.com/orcaman/concurrent-map

zerolog 
(c) 2017 Olivier Poitrey
License (MIT): github.com/rs/zerolog

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
