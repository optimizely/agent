# Getting Started Guide

## Installation
Optimizely Agent is available via most application package managers and can be installed with a single command:

### Via RPM (CentOS)
```bash
> sudo yum install optimizely
> sudo service optimizely start
```

### Via DEB (Ubuntu)
```bash
> sudo apt-get install optimizely
> sudo service optimizely start
```

### Via Homebrew (OSX)
```bash
> brew install optimizely
> brew services start optimizely
```

Once installed and the service is running we can start to explore the REST APIs.

## Evaluating REST APIs
The rest of the getting started guide will demonstrate the APIs capabilities. For brevity, we've chosen to illustrate the API usage with Python. Note that the APIs are also defined via OpenAPI (Swagger) and can be found [here](http://localhost:8080/openapi.yaml).

### Start an http session
Each request made into Optimizely Agent is in the context of an Optimizely SDK Key. SDK Keys map API requests to a specific Optimizely Project and Environment. We can setup a global request header by using the `requests.Session` object.

```python
import requests

s = requests.Session()
s.headers.update({'X-Optimizely-SDK-Key': '<<YOUR-SDK-KEY>>'})
```

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
The `/activate?featureKey={key}` endpoint activates the feature for a given user. In Optimizely, activation is in the context of a given user to make the relative bucketing decision. In this case we'll provide a `userId` via the request body. The `userId` will be used to determine how the feature feature will be evaluated. Features can either be part of a Feature Test in which variations of feature variables are being measured against one another or a feature rollout, which progressively make the feature available to the selected audience.

From an API standpoint the presence of a Feature Test or Rollout is abstrated away from the response and only the resulting variation or enabled feature is returned.

```python
params = { "featureKey": "my-feature" }
payload = { "userId": "test-user" }
resp = s.post(url = 'http://localhost:8080/v1/activate', params=params, json=payload)

print(resp.json())
```

The activate API is a POST to signal to the caller that there are side-effects. Namely, activation results in a "decision" event sent to Optimizely analytics for the purpose of analyzing Feature Test results. A "decision" will NOT be sent if the feature is simply part of a rollout.
