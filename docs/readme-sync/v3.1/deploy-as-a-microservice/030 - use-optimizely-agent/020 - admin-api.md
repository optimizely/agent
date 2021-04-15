---
title: "Admin API"
excerpt: ""
slug: "admin-api"
hidden: false
metadata: 
  title: "Admin APIs - Optimizely Full Stack"
createdAt: "2020-02-21T17:44:28.054Z"
updatedAt: "2020-02-21T23:09:19.274Z"
---
The Admin API provides system information about the running process. This can be used to check the availability of the service, runtime information and operational metrics. By default the admin listener is configured on port 8088.

## Info

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

## Health Check

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

## Metrics

The `/metrics` endpoint exposes telemetry data of the running Optimizely Agent. The core runtime metrics are exposed via the go expvar package. Documentation for the various statistics can be found as part of the [mstats](https://golang.org/src/runtime/mstats.go) package.

Example Request:
```bash
curl localhost:8088/metrics
```

Example Response:
```json
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

```bash
"timers.<metric-name>.counts": 0,
"timers.<metric-name>.responseTime": 0,
"timers.<metric-name>.responseTimeHist.p50": 0,
"timers.<metric-name>.responseTimeHist.p90": 0,
"timers.<metric-name>.responseTimeHist.p95": 0,
"timers.<metric-name>.responseTimeHist.p99": 0,
```