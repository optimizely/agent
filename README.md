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
Optimizely Agent is implemented in [Golang](https://golang.org/). Golang version 1.13+ is required for developing and compiling from source.
Installers and binary archives for most platforms can be downloaded directly from the Go [downloads](https://golang.org/dl/) page.

## Running Optimizely from source
Once Go is installed, the Optimizely Agent can be started via the following `make` command:
```bash
make run
```
This will start the Optimizely Agent with the default configuration in the foreground.

## Running Optimizely via Docker
Alternatively, if you have Docker installed, Optimizely Agent can be started as a container:
```bash
docker run -d --name optimizely-agent \
         -p 8080:8080 \
         -p 8088:8088 \
         -p 8085:8085 \
         --env OPTIMIZELY_LOG_PRETTY=true \
         optimizely/agent:latest
```
The above command also shows how environment variables can be passed in to alter the configuration without having to
create a config.yaml file. See the [configuration](#configuration-options) for more options.

When a new version is released, 2 images are pushed to dockerhub, they are distinguished by their tags:
- :latest (same as :X.Y.Z)
- :alpine (same as :X.Y.Z-alpine)

The difference between latest and alpine is that latest is built `FROM scratch` while alpine is `FROM alpine`.
- [latest Dockerfile](./scripts/dockerfiles/Dockerfile.static)
- [alpine Dockerfile](./scripts/dockerfiles/Dockerfile.alpine)

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

## Full Stack API

The core Full Stack API is implemented as a REST service configured on it's own HTTP listener port (default 8080).
The full API specification is defined in an OpenAPI 3.0 (aka Swagger) [spec](./api/openapi-spec/openapi.yaml).

Each request made into the Full Stack API must include a `X-Optimizely-SDK-Key` in the request header to
identify the context the request should be evaluated. The SDK key maps to a unique Optimizely Project and
[Environment](https://docs.developers.optimizely.com/rollouts/docs/manage-environments) allowing multiple
Environments to be serviced by a single Agent.

## Webhooks

The webhook listener used to receive inbound [Webhook](https://docs.developers.optimizely.com/rollouts/docs/webhooks)
requests from optimizely.com. These webhooks enable PUSH style notifications triggering immediate project configuration updates.
The webhook listener is configured on its own port (default: 8085) since it can be configured to select traffic from the internet.

To accept webhook requests Agent must be configured by mapping an Optimizely Project Id to a set of SDK keys along
with the associated secret used for validating the inbound request. An example webhook configuration can
be found in the the provided [config.yaml](./config.yaml).

## Admin API

The Admin API provides system information about the running process. This can be used to check the availability
of the service, runtime information and operational metrics. By default the admin listener is configured on port 8088.

### Info

The `/info` endpoint provides basic information about the Optimizely Agent instance.

Example Request:
```bash
curl localhost:8088/info
```

Example Response:
```json
{
    "version": "v0.10.0",
    "author": "Optimizely Inc.",
    "app_name": "optimizely"
}
```

### Health Check

The `/health` endpoint is used to determine service availability.

Example Request:
```bash
curl localhost:8088/health
```

Example Response:
```json
{
    "status": "ok"
}
```

Agent will return a HTTP 200 - OK response if and only if all configured listeners are open and all external dependent services can be reached.
A non-healthy service will return a HTTP 503 - Unavailable response with a descriptive message to help diagnose the issue.

This endpoint can used when placing Agent behind a load balancer to indicate whether a particular instance can receive inbound requests.

### Metrics

The `/metrics` endpoint exposes telemetry data of the running Optimizely Agent. The core runtime metrics are exposed via the go expvar package. Documentation for the various statistics can be found as part of the [mstats](https://golang.org/src/runtime/mstats.go) package.

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
        "HeapAlloc": 924136,
        ...
        "Frees": 172
    },
    ...
}
```
Custom metrics are also provided for the individual service endpoints and follow the pattern of:

```
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
