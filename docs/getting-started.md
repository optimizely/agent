# Getting Started Guide

## Installation
Sidedoor is available via most application package managers and can be installed with a single command:

### Via RPM (CentOS)
```bash
> sudo yum install sidedoor
> sudo service sidedoor start
```

### Via DEB (Ubuntu)
```bash
> sudo apt-get install sidedoor
> sudo service sidedoor start
```

### Via Homebrew (OSX)
```bash
> brew install sidedoor
> brew services start sidedoor
```

Once installed and the service is running we can start to explore the REST APIs.

## Evaluating REST APIs
The rest of the getting started guide will demonstrate the APIs capabilities. For brevity, we've chosen to illustrate the API usage with Python. Note that the APIs are also defined via OpenAPI (Swagger) and can be found [here](http://localhost:8080/openapi.yaml).

### Start an http session
Each request made into Sidedoor is in the context of an Optimizely SDK Key. SDK Keys map API requests to a specific Optimizely Project and Environment. We can setup a global request header by using the `requests.Session` object.

```python
import requests

s = requests.Session()
s.headers.update({'X-Optimizely-SDK-Key': '<<YOUR-SDK-KEY>>'})
```

Future examples will assume this session is being maintained.

### List Features
The `/features` endpoint returns a list of all available features.

```python
resp = s.get('http://localhost:8080/features')
features = resp.json()

for feature in features:
    print(feature['Key'])
```

### Get Feature
The `/features/:key` endpoint returns the feature associated with the key provided in the path parameter.

```python
feature_key = 'feature-key'
resp = s.get('http://localhost:8080/features/{}'.format(feature_key))

print(resp.json())
```

### Activate Feature
The `/features/:key/activate` endpoint activates the feature for a given user. In Optimizely, activation is in the context of a given user. In this case we'll provide a `userId` via a URL parameter. The `userId` will be used to determine how the feature feature will be returned, if at all. Features can either be part of a Feature Test by which multiple variations of a feature are being measured against one another and a feature rollout, which is a percentage based membership criteria. (TODO reword).

From an API standpoint the presence of a Feature Test or Rollout is abstrated away from the response and only the resulting variation or enabled feature is returned.

```python
s.params.update({'userId': 'test-user'})
resp = s.POST('http://localhost:8080/features/{}/activate'.format(feature_key))

print(resp.json())
```

The activate API is a POST to signal to the caller that there are side-effects. Namely, activation results in a "decision" event sent to Optimizely analytics for the purpose of analyzing Feature Test results. A "decision" will NOT be sent if the feature is simply parat of a rollout.
