---
title: "Quickstart with Docker"
excerpt: ""
slug: "quickstart-with-docker"
hidden: true
metadata: 
  title: "Agent Quickstart with Docker - Optimizely Full Stack"
createdAt: "2020-05-21T20:35:58.387Z"
updatedAt: "2020-07-14T20:51:52.458Z"
---

This is a brief quickstart showing how to run Agent locally via Docker and how to access some of the common API endpoints.
If Docker is not installed then you can download it [here](https://docs.docker.com/install/).

## Running locally via Docker
First pull the Docker image with:
```bash
docker pull optimizely/agent
```

Then start the service in the foreground with the following command:
```bash
docker run -p 8080:8080 --env OPTIMIZELY_LOG_PRETTY=true optimizely/agent
```
Note that we're enabling "pretty" logs which provide colorized and human readable formatting.
The default log output format is structured JSON. 

## Evaluating REST APIs
The rest of the getting started guide will demonstrate the APIs capabilities. For brevity, we've chosen to illustrate the API usage with Python. Note that the APIs are also defined via OpenAPI (Swagger) and can be found [here](http://localhost:8080/openapi.yaml).

### Start an http session
Each request made into Optimizely Agent is in the context of an Optimizely SDK Key. SDK Keys map API requests to a specific Optimizely Project and Environment. We can setup a global request header by using the `requests.Session` object.

```python
import requests

s = requests.Session()
s.headers.update({'X-Optimizely-SDK-Key': '<<YOUR-SDK-KEY>>'})
```

To get your SDK key, navigate to the project settings of your Optimizely account.

Future examples will assume this session is being maintained.

### Get current environment configuration
The `/config` endpoint returns a manifest of the current working environment.

```python
resp = s.get('http://localhost:8080/v1/config')
env = resp.json()

for key in env['featuresMap']:
    print(key)
```

### Activate Feature
The `/activate?featureKey={key}` endpoint activates the feature for a given user. In Optimizely, activation is in the context of a given user to make the relative bucketing decision. In this case we'll provide a `userId` via the request body. The `userId` will be used to determine how the feature will be evaluated. Features can either be part of a Feature Test in which variations of feature variables are being measured against one another or a feature rollout, which progressively make the feature available to the selected audience.

From an API standpoint the presence of a Feature Test or Rollout is abstracted away from the response and only the resulting variation or enabled feature is returned.

```python
params = { "featureKey": "my-feature" }
payload = { "userId": "test-user" }
resp = s.post(url = 'http://localhost:8080/v1/activate', params=params, json=payload)

print(resp.json())
```

The activate API is a POST to signal to the caller that there are side-effects. Namely, activation results in a "decision" event sent to Optimizely analytics for the purpose of analyzing Feature Test results. A "decision" will NOT be sent if the feature is simply part of a rollout. 
