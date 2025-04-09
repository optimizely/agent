[![Build Status](https://github.com/optimizely/agent/actions/workflows/agent.yml/badge.svg?branch=master)](https://github.com/optimizely/agent/actions/workflows/agent.yml?query=branch%3Amaster)
[![Coverage Status](https://coveralls.io/repos/github/optimizely/agent/badge.svg)](https://coveralls.io/github/optimizely/agent)

# Optimizely Agent

This repository houses the Optimizely Agent service for use with Optimizely Feature Experimentation and Optimizely Full Stack (legacy).

Optimizely Feature Experimentation is an A/B testing and feature management tool for product development teams that enables you to experiment at every step. Using Optimizely Feature Experimentation allows for every feature on your roadmap to be an opportunity to discover hidden insights. Learn more at [Optimizely.com](https://www.optimizely.com/products/experiment/feature-experimentation/), or see the [developer documentation](https://docs.developers.optimizely.com/experimentation/v4.0.0-full-stack/docs/welcome).

Optimizely Rollouts is [free feature flags](https://www.optimizely.com/free-feature-flagging/) for development teams. You can easily roll out and roll back features in any application without code deploys, mitigating risk for every feature on your roadmap.

## Get Started

Refer to the [Agent's developer documentation](https://docs.developers.optimizely.com/experimentation/v4.0.0-full-stack/docs/optimizely-agent) for detailed instructions on getting started with using the SDK.

### Requirements

Optimizely Agent is implemented in [Golang](https://golang.org/). Golang version 1.21.0+ is required for developing and compiling from source.
Installers and binary archives for most platforms can be downloaded directly from the Go [downloads](https://go.dev/dl/) page.

### Run from source (Linux / OSX)

Once Go is installed, the Optimizely Agent can be started via the following `make` command:

```bash
make setup
make run
```

This will start the Optimizely Agent with the default configuration in the foreground.

### Run from source (Windows)

A helper script is available under [scripts/build.ps1](./scripts/build.ps1) to automate compiling Agent in a Windows environment. The script will download and install both Git and Golang and then attempt to compile Agent. Open a Powershell terminal and run

```bash
Set-ExecutionPolicy -ExecutionPolicy Unrestricted -Scope CurrentUser

.\scripts\build.ps1

.\bin\optimizely.exe
```

### Run via Docker

If you have Docker installed, Optimizely Agent can be started as a container. First pull the Docker image with:

```bash
docker pull optimizely/agent
```

By default this will pull the "latest" tag. You can also specify a specific version of Agent by providing the version
as a tag to the docker command:

```bash
docker pull optimizely/agent:X.Y.Z
```

Then run the docker container with:

```bash
docker run -p 8080:8080 --env OPTIMIZELY_LOG_PRETTY=true --env OPTIMIZELY_SERVER_HOST=0.0.0.0 --env OPTIMIZELY_SERVER_ALLOWEDHOSTS=127.0.0.1 optimizely/agent
```

This will start Agent in the foreground and expose the container API port 8080 to the host.

Note that when a new version is released, 2 images are pushed to dockerhub, they are distinguished by their tags:

- :latest (same as :X.Y.Z)
- :alpine (same as :X.Y.Z-alpine)

The difference between latest and alpine is that latest is built `FROM scratch` while alpine is `FROM alpine`.

- [latest Dockerfile](./scripts/dockerfiles/Dockerfile.static)
- [alpine Dockerfile](./scripts/dockerfiles/Dockerfile.alpine)

### Configuration Options

Optimizely Agent configuration can be overridden by a yaml configuration file provided at runtime.

By default the configuration file will be sourced from the current active directory `e.g. ./config.yaml`.
Alternative configuration locations can be specified at runtime via environment variable or command line flag.

```bash
OPTIMIZELY_CONFIG_FILENAME=config.yaml make run
```

The default configuration can be found [here](config.yaml).

Below is a comprehensive list of available configuration properties.

| Property Name                                     | Env Variable                                    | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| ------------------------------------------------- | ----------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| admin.auth.clients                                | N/A                                             | Credentials for requesting access tokens. See: [Authorization Guide](https://docs.developers.optimizely.com/experimentation/v4.0.0-full-stack/docs/authorization)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| admin.auth.hmacSecrets                            | OPTIMIZELY_ADMIN_AUTH_HMACSECRETS               | Signing secret for issued access tokens. See: [Authorization Guide](https://docs.developers.optimizely.com/experimentation/v4.0.0-full-stack/docs/authorization)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| admin.auth.jwksUpdateInterval                     | OPTIMIZELY_ADMIN_AUTH_JWKSUPDATEINTERVAL        | JWKS Update Interval for caching the keys in the background. See: [Authorization Guide](https://docs.developers.optimizely.com/experimentation/v4.0.0-full-stack/docs/authorization)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| admin.auth.jwksURL                                | OPTIMIZELY_ADMIN_AUTH_JWKSURL                   | JWKS URL for validating access tokens. See: [Authorization Guide](https://docs.developers.optimizely.com/experimentation/v4.0.0-full-stack/docs/authorization)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| admin.auth.ttl                                    | OPTIMIZELY_ADMIN_AUTH_TTL                       | Time-to-live of issued access tokens. See: [Authorization Guide](https://docs.developers.optimizely.com/experimentation/v4.0.0-full-stack/docs/authorization)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| admin.port                                        | OPTIMIZELY_ADMIN_PORT                           | Admin listener port. Default: 8088                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| api.auth.clients                                  | N/A                                             | Credentials for requesting access tokens. See: [Authorization Guide](https://docs.developers.optimizely.com/experimentation/v4.0.0-full-stack/docs/authorization)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| api.auth.hmacSecrets                              | OPTIMIZELY_API_AUTH_HMACSECRETS                 | Signing secret for issued access tokens. See: [Authorization Guide](https://docs.developers.optimizely.com/experimentation/v4.0.0-full-stack/docs/authorization)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| api.auth.jwksUpdateInterval                       | OPTIMIZELY_API_AUTH_JWKSUPDATEINTERVAL          | JWKS Update Interval for caching the keys in the background. See: [Authorization Guide](https://docs.developers.optimizely.com/experimentation/v4.0.0-full-stack/docs/authorization)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| api.auth.jwksURL                                  | OPTIMIZELY_API_AUTH_JWKSURL                     | JWKS URL for validating access tokens. See: [Authorization Guide](https://docs.developers.optimizely.com/experimentation/v4.0.0-full-stack/docs/authorization)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| api.auth.ttl                                      | OPTIMIZELY_API_AUTH_TTL                         | Time-to-live of issued access tokens. See: [Authorization Guide](https://docs.developers.optimizely.com/experimentation/v4.0.0-full-stack/docs/authorization)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| api.enableNotifications                           | OPTIMIZELY_API_ENABLENOTIFICATIONS              | Enable streaming notification endpoint. Default: false                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| api.enableOverrides                               | OPTIMIZELY_API_ENABLEOVERRIDES                  | Enable bucketing overrides endpoint. Default: false                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| api.maxConns                                      | OPTIMIZELY_API_MAXCONNS                         | Maximum number of concurrent requests                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| api.port                                          | OPTIMIZELY_API_PORT                             | Api listener port. Default: 8080                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| author                                            | OPTIMIZELY_AUTHOR                               | Agent author. Default: Optimizely Inc.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| client.batchSize                                  | OPTIMIZELY_CLIENT_BATCHSIZE                     | The number of events in a batch. Default: 10                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| client.datafileURLTemplate                        | OPTIMIZELY_CLIENT_DATAFILEURLTEMPLATE           | Template URL for SDK datafile location. Default: https://cdn.optimizely.com/datafiles/%s.json                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| client.eventURL                                   | OPTIMIZELY_CLIENT_EVENTURL                      | URL for dispatching events. Default: https://logx.optimizely.com/v1/events                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| client.flushInterval                              | OPTIMIZELY_CLIENT_FLUSHINTERVAL                 | The maximum time between events being dispatched. Default: 30s                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| client.pollingInterval                            | OPTIMIZELY_CLIENT_POLLINGINTERVAL               | The time between successive polls for updated project configuration. Default: 1m                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| client.queueSize                                  | OPTIMIZELY_CLIENT_QUEUESIZE                     | The max number of events pending dispatch. Default: 1000                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| client.sdkKeyRegex                                | OPTIMIZELY_CLIENT_SDKKEYREGEX                   | Regex to validate SDK keys provided in request header. Default: ^\\w+(:\\w+)?$                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| client.userProfileService                         | OPTIMIZELY_CLIENT_USERPROFILESERVICE            | Property used to enable and set UserProfileServices. Default: ./config.yaml                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| client.odp.disable                                | OPTIMIZELY_CLIENT_ODP_DISABLE                   | Property used to disable odp. Default: false                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| client.odp.eventsRequestTimeout                   | OPTIMIZELY_CLIENT_ODP_EVENTSREQUESTTIMEOUT      | Property used to update timeout in seconds after which event requests will timeout. Default: 10s                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| client.odp.eventsFlushInterval                    | OPTIMIZELY_CLIENT_ODP_EVENTSFLUSHINTERVAL       | Property used to update flush interval in seconds for events. Default: 1s                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| client.odp.segmentsRequestTimeout                 | OPTIMIZELY_CLIENT_ODP_SEGMENTSREQUESTTIMEOUT    | Property used to update timeout in seconds after which segment requests will timeout: 10s                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| client.odp.cache                                  | OPTIMIZELY_CLIENT_ODP_SEGMENTSCACHE             | Property used to enable and set cache service for odp. Default: ./config.yaml                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| config.filename                                   | OPTIMIZELY_CONFIG_FILENAME                      | Location of the configuration YAML file. Default: ./config.yaml                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| log.level                                         | OPTIMIZELY_LOG_LEVEL                            | The log [level](https://github.com/rs/zerolog#leveled-logging) for the agent. Default: info                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| log.pretty                                        | OPTIMIZELY_LOG_PRETTY                           | Flag used to set colorized console output as opposed to structured json logs. Default: false                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| name                                              | OPTIMIZELY_NAME                                 | Agent name. Default: optimizely                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| sdkKeys                                           | OPTIMIZELY_SDKKEYS                              | Comma delimited list of SDK keys used to initialize on startup                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| server.allowedHosts                               | OPTIMIZELY_SERVER_ALLOWEDHOSTS                  | List of allowed request host values. Requests whose host value does not match either the configured server.host, or one of these, will be rejected with a 404 response. To match all subdomains, you can use a leading dot (for example `.example.com` matches `my.example.com`, `hello.world.example.com`, etc.). You can use the value `.` to disable allowed host checking, allowing requests with any host. Request host is determined in the following priority order: 1. X-Forwarded-Host header value, 2. Forwarded header host= directive value, 3. Host property of request (see Host under https://pkg.go.dev/net/http#Request). Note: don't include port in these hosts values - port is stripped from the request host before comparing against these. |
| server.batchRequests.maxConcurrency               | OPTIMIZELY_SERVER_BATCHREQUESTS_MAXCONCURRENCY  | Number of requests running in parallel. Default: 10                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| server.batchRequests.operationsLimit              | OPTIMIZELY_SERVER_BATCHREQUESTS_OPERATIONSLIMIT | Number of allowed operations. ( will flag an error if the number of operations exeeds this parameter) Default: 500                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| server.certfile                                   | OPTIMIZELY_SERVER_CERTFILE                      | Path to a certificate file, used to run Agent with HTTPS                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| server.disabledCiphers                            | OPTIMIZELY_SERVER_DISABLEDCIPHERS               | List of TLS ciphers to disable when accepting HTTPS connections                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| server.healthCheckPath                            | OPTIMIZELY_SERVER_HEALTHCHECKPATH               | Path for the health status api. Default: /health                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| server.host                                       | OPTIMIZELY_SERVER_HOST                          | Host of server. Default: 127.0.0.1                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| server.interceptors                               | N/A                                             | Property used to enable and set [Interceptor](https://docs.developers.optimizely.com/experimentation/v4.0.0-full-stack/docs/agent-plugins#interceptor-plugins) plugins                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| server.keyfile                                    | OPTIMIZELY_SERVER_KEYFILE                       | Path to a key file, used to run Agent with HTTPS                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| server.readTimeout                                | OPTIMIZELY_SERVER_READTIMEOUT                   | The maximum duration for reading the entire body. Default: “5s”                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| server.writeTimeout                               | OPTIMIZELY_SERVER_WRITETIMEOUT                  | The maximum duration before timing out writes of the response. Default: “10s”                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| version                                           | OPTIMIZELY_VERSION                              | Agent version. Default: `git describe --tags`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| webhook.port                                      | OPTIMIZELY_WEBHOOK_PORT                         | Webhook listener port: Default: 8085                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| webhook.projects.<_projectId_>.sdkKeys            | N/A                                             | Comma delimited list of SDK Keys applicable to the respective projectId                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| webhook.projects.<_projectId_>.secret             | N/A                                             | Webhook secret used to validate webhook requests originating from the respective projectId                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| webhook.projects.<_projectId_>.skipSignatureCheck | N/A                                             | Boolean to indicate whether the signature should be validated. TODO remove in favor of empty secret.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |

More information about configuring Agent can be found in the [Advanced Configuration Notes](https://docs.developers.optimizely.com/experimentation/v4.0.0-full-stack/docs/advanced-configuration).

### API

The core API is implemented as a REST service configured on it's own HTTP listener port (default 8080).
The full API specification is defined in an OpenAPI 3.0 (aka Swagger) [spec](./api/openapi-spec/openapi.yaml).

Each request made into the API must include a `X-Optimizely-SDK-Key` in the request header to
identify the context the request should be evaluated. The SDK key maps to a unique Optimizely Project and
[Environment](https://docs.developers.optimizely.com/experimentation/v4.0.0-full-stack/docs/manage-environments) allowing multiple
Environments to be serviced by a single Agent.
For a secure environment, this header should include both the SDK key and the datafile access token
separated by a colon. For example, if SDK key is `my_key` and datafile access token is `my_token`
then set header's value to `my_key:my_token`.

#### Enabling CORS

CORS can be enabled for the core API service by setting the the appropriate cors properties.

```yaml
api:
  cors:
    allowedMethods:
      - "HEAD"
      - "GET"
      - "POST"
      - "OPTIONS"
```

For more advanced options please refer to the [go-chi/cors](https://github.com/go-chi/cors) middleware documentation.

NOTE: To avoid any potential security issues, and reduce risk to your data it's recommended that [authentication](https://docs.developers.optimizely.com/experimentation/v4.0.0-full-stack/docs/authorization)
is enabled alongside CORS.

### Webhooks

The webhook listener used to receive inbound [Webhook](https://docs.developers.optimizely.com/experimentation/v4.0.0-full-stack/docs/webhooks-agent)
requests from optimizely.com. These webhooks enable PUSH style notifications triggering immediate project configuration updates.
The webhook listener is configured on its own port (default: 8085) since it can be configured to select traffic from the internet.

To accept webhook requests Agent must be configured by mapping an Optimizely Project Id to a set of SDK keys along
with the associated secret used for validating the inbound request. An example webhook configuration can
be found in the the provided [config.yaml](./config.yaml).

When running Agent in High Availability (HA) mode, it's important to ensure that all nodes are updated promptly when a webhook event (datafile updated) is received. By default, only one Agent node or instance will receive the webhook notification. A pub-sub system can be used to ensure this.

Redis, a powerful in-memory data structure store, can be used as a relay to propagate the datafile webhook event to all other nodes in the HA setup. This ensures that all nodes are notified about the event and can update their datafiles accordingly.

To set up Redis as a relay, you need to enable the datafile synchronization in your Optimizely Agent configuration. The PubSub feature of Redis is used to publish the webhook notifications to all subscribed Agent nodes.

Here's an example of how you can enable the datafile synchronization with Redis:

```yaml
## synchronization should be enabled when features for multiple nodes like notification streaming are deployed
synchronization:
    pubsub:
        redis:
            host: "localhost:6379"
            password: ""
            database: 0
    ## if datafile synchronization is enabled, then for each webhook API call
    ## the datafile will be sent to all available replicas to achieve better eventual consistency
    datafile:
        enable: true
        default: "redis"
```

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

The `/metrics` endpoint exposes telemetry data of the running Optimizely Agent.

Currently, Agent exposes two type of metrics data (expvar or prometheus) based on user's input. By default, expvar metrics will be used. To configure this, update config.yaml or
set the value of the environment variable `OPTIMIZELY_ADMIN_METRICSTYPE`. Supported values are `expvar` (default) & `promethues`.

```yaml
##
## admin service configuration
##
admin:
    ## http listener port
    port: "8088"
    ## metrics package to use
    ## supported packages are expvar and prometheus
    ## default is expvar
    metricsType: ""
    ## metricsType: "promethues" ## for prometheus metrics
```

#### Expvar metrics

The core runtime metrics are exposed via the Go expvar package. Documentation for the various statistics can be found as part of the [mstats](https://go.dev/src/runtime/mstats.go) package.

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

#### Prometheus metrics

Optimizely Agent also supports Prometheus metrics. Prometheus is an open-source toolkit for monitoring and alerting. You can use it to collect and visualize metrics in a time-series database.

To access the Prometheus metrics, you can use the `/metrics` endpoint with a Prometheus server. The metrics are exposed in a format that Prometheus can scrape and aggregate.

Example Request:

```bash
curl localhost:8088/metrics
```

This will return a plain text response in the Prometheus Exposition Format, which includes all the metrics that Prometheus is currently tracking.

Please note that you need to configure your Prometheus server to scrape metrics from this endpoint.

For more information on how to use Prometheus for monitoring, you can refer to the [official Prometheus documentation](https://prometheus.io/docs/introduction/overview/).

Example Response:

```
...
# HELP promhttp_metric_handler_requests_total Total number of scrapes by HTTP status code.
# TYPE promhttp_metric_handler_requests_total counter
promhttp_metric_handler_requests_total{code="200"} 1
promhttp_metric_handler_requests_total{code="500"} 0
promhttp_metric_handler_requests_total{code="503"} 0
# HELP timer_decide_hits 
# TYPE timer_decide_hits counter
timer_decide_hits 1
# HELP timer_decide_response_time 
# TYPE timer_decide_response_time counter
timer_decide_response_time 658.109
# HELP timer_decide_response_time_hist 
# TYPE timer_decide_response_time_hist histogram
timer_decide_response_time_hist_bucket{le="0.005"} 0
timer_decide_response_time_hist_bucket{le="0.01"} 0
timer_decide_response_time_hist_bucket{le="0.025"} 0
timer_decide_response_time_hist_bucket{le="0.05"} 0
timer_decide_response_time_hist_bucket{le="0.1"} 0
timer_decide_response_time_hist_bucket{le="0.25"} 0
timer_decide_response_time_hist_bucket{le="0.5"} 0
timer_decide_response_time_hist_bucket{le="1"} 0
timer_decide_response_time_hist_bucket{le="2.5"} 0
timer_decide_response_time_hist_bucket{le="5"} 0
timer_decide_response_time_hist_bucket{le="10"} 0
timer_decide_response_time_hist_bucket{le="+Inf"} 1
timer_decide_response_time_hist_sum 658.109
timer_decide_response_time_hist_count 1
# HELP timer_track_event_hits 
# TYPE timer_track_event_hits counter
timer_track_event_hits 1
# HELP timer_track_event_response_time 
# TYPE timer_track_event_response_time counter
timer_track_event_response_time 0.356334
# HELP timer_track_event_response_time_hist 
# TYPE timer_track_event_response_time_hist histogram
timer_track_event_response_time_hist_bucket{le="0.005"} 0
timer_track_event_response_time_hist_bucket{le="0.01"} 0
timer_track_event_response_time_hist_bucket{le="0.025"} 0
timer_track_event_response_time_hist_bucket{le="0.05"} 0
timer_track_event_response_time_hist_bucket{le="0.1"} 0
timer_track_event_response_time_hist_bucket{le="0.25"} 0
timer_track_event_response_time_hist_bucket{le="0.5"} 1
timer_track_event_response_time_hist_bucket{le="1"} 1
timer_track_event_response_time_hist_bucket{le="2.5"} 1
timer_track_event_response_time_hist_bucket{le="5"} 1
timer_track_event_response_time_hist_bucket{le="10"} 1
timer_track_event_response_time_hist_bucket{le="+Inf"} 1
timer_track_event_response_time_hist_sum 0.356334
timer_track_event_response_time_hist_count 1
...
```

### Profiling

Agent exposes the runtime profiling data in the format expected by the [pprof](https://github.com/google/pprof/blob/master/doc/README.md) visualization tool.

You can use the pprof tool to look at the heap profile:

```
go tool pprof http://localhost:6060/debug/pprof/heap
```

Or to look at a 5-second CPU profile: (higher durations require configuring the `server.writeTimeout`)

```
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=5
```

Or to look at the goroutine blocking profile, after setting `runtime.blockProfileRate` in the configuration:

```
go tool pprof http://localhost:8088/debug/pprof/block
```

Or to collect a 5-second execution trace:

```
wget "http://localhost:8088/debug/pprof/trace?seconds=5"
```

Or to look at the holders of contended mutexes, after setting `runtime.mutexProfileFraction` in your configuration:

```
go tool pprof http://localhost:6060/debug/pprof/mutex
```

To view all available profiles can be found at [http://localhost:8088/debug/pprof/](http://localhost:8088/debug/pprof/) in your browser.

## Agent Plugins

Optimizely Agent can be extended through the use of [plugins](https://docs.developers.optimizely.com/experimentation/v4.0.0-full-stack/docs/agent-plugins). Plugins are distinct from the standard Agent packages
to provide a namespaced environment for custom logic. Plugins must be compiled as part of the Agent distribution and are enabled through configuration.

### Interceptor Plugins

- [httplog](./plugins/interceptors/httplog) - Adds HTTP request logging based on [go-chi/httplog](https://github.com/go-chi/httplog).

### UserProfileService Plugins

- [UserProfileService](./plugins/userprofileservice/README.md) - Adds UserProfileService.

### ODPCache Plugins

- [ODPCache](./plugins/odpcache/README.md) - Adds ODP Cache.

### Authorization

Optimizely Agent supports authorization workflows based on OAuth and JWT standards, allowing you to protect access to its API and Admin interfaces. For details, see [Authorization Guide](https://docs.developers.optimizely.com/experimentation/v4.0.0-full-stack/docs/authorization).

### Notifications

Just as you can use Notification Listeners to subscribe to events of interest with Optimizely SDKs, you can use the Notifications endpoint to subscribe to events in Agent. For more information, see the [Notifications Guide](https://docs.developers.optimizely.com/experimentation/v4.0.0-full-stack/docs/agent-notifications).

When the Agent is operating in High Availability (HA) mode, you need to enable notification synchronization to get notifications from all nodes in an HA setup. A PubSub system (Redis) is now used to ensure consistent retrieval of notification events across all nodes in an HA setup.
Here's an example of how you can enable the notification synchronization with Redis:

```yaml
## synchronization should be enabled when features for multiple nodes like notification streaming are deployed
synchronization:
    pubsub:
        redis:
            host: "localhost:6379"
            password: ""
            database: 0
    ## if notification synchronization is enabled, then the active notification event-stream API
    ## will get the notifications from available replicas
    notification:
        enable: true
        default: "redis"
```

## Agent Development

### Package Structure

Following best practice for go project layout as defined [here](https://github.com/golang-standards/project-layout)

- **api** - OpenAPI/Swagger specs, JSON schema files, protocol definition files.
- **bin** - Compiled application binaries.
- **cmd** - Main applications for this project.
- **config** - Application configuration.
- **docs** - User documentation files.
- **pkg** - Library code that can be used by other applications.
- **plugins** - Plugin libraries for extending Agent functionality.
- **scripts** - Scripts to perform various build, install, analysis, etc operations.

### Make Commands

The following `make` targets can be used to build and run the application:

- **build** - builds optimizely and installs binary in bin/optimizely
- **clean** - runs `go clean` and removes the bin/ dir
- **cover** - runs test suite with coverage profiling
- **cover-html** - generates test coverage html report
- **setup** - installs all dev and ci dependencies, but does not install golang
- **lint** - runs `golangci-lint` linters defined in `.golangci.yml` file
- **run** - builds and executes the optimizely binary
- **test** - recursively tests all .go files

## Credits

This software is used with additional code that is separately downloaded by you. These components are subject to their own license terms which you should review carefully.

Gohistogram
(c) 2013 VividCortex
License (MIT): https://github.com/VividCortex/gohistogram

Chi
(c) 2015-present Peter Kieltyka (https://github.com/pkieltyka), Google Inc.
License (MIT): https://github.com/go-chi/chi

chi-render
(c) 2016-Present https://github.com/go-chi ‑ authors
License (MIT): https://github.com/go-chi/render

hostrouter
(c) 2016-Present https://github.com/go-chi - authors
License (MIT): https://github.com/go-chi/hostrouter

go-chi/cors
(c) 2014 Olivier Poitrey
License (MIT): https://github.com/go-chi/cors

go-kit
(c) 2015 Peter Bourgon
License (MIT): https://github.com/go-kit/kit

guuid
(c) 2009,2014 Google Inc. All rights reserved.
License (BSD 3-Clause): https://github.com/google/uuid

optimizely go sdk
(c) 2016-2017, Optimizely, Inc. and contributors
License (Apache 2): https://github.com/optimizely/go-sdk

concurrent-map
(c) 2014 streamrail
License (MIT): https://github.com/orcaman/concurrent-map

zerolog
(c) 2017 Olivier Poitrey
License (MIT): https://github.com/rs/zerolog

viper
(c) 2014 Steve Francia
License (MIT): https://github.com/spf13/viper

testify
(c) 2012-2018 Mat Ryer and Tyler Bunnell
License (MIT): https://github.com/stretchr/testify

net
(c) 2009 The Go Authors
License (BSD 3-Clause): https://github.com/golang/net

sync
(c) 2009 The Go Authors
License (BSD 3-Clause): https://github.com/golang/sync

statik
(c) 2014 rakyll
License (Apache 2): https://github.com/rakyll/statik v0.1.7

sys
(c) 2009 The Go Authors
License (BSD 3-Clause): https://github.com/golang/sys

opentelemetry-go
Copyright The OpenTelemetry Authors
License (Apache-2.0): https://github.com/open-telemetry/opentelemetry-go

prometheus client_golang
Copyright 2015 The Prometheus Authors
License (Apache-2.0): https://github.com/prometheus/client_golang

## Apache Copyright Notice

Copyright 2019-present, Optimizely, Inc. and contributors

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

test
