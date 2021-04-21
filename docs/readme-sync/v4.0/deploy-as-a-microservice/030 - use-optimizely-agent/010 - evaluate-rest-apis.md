---
title: "Evaluate REST APIs"
excerpt: ""
slug: "evaluate-rest-apis"
hidden: false
metadata: 
  title: "Evaluate REST APIs - Optimizely Full Stack"
createdAt: "2020-02-21T17:44:53.019Z"
updatedAt: "2021-03-15T23:02:34.056Z"
---
Below is an example demonstrating the APIs capabilities. For brevity, we've chosen to illustrate the API usage with Python. Note that the API documentation is defined via an OpenAPI (Swagger) spec and can be viewed [here](https://library.optimizely.com/docs/api/agent/v1/index.html).

## Start an http session
Each request made into Optimizely Agent is in the context of an Optimizely SDK Key. SDK Keys map API requests to a specific Optimizely Project and Environment. We can setup a global request header by using the `requests.Session` object.


```python
import requests

s = requests.Session()
s.headers.update({'X-Optimizely-SDK-Key': 'YOUR-SDK-KEY'})
```
The following examples will assume this session is being maintained.

## Get current environment configuration
The `/v1/config` endpoint returns a manifest of the current working environment.

```python
resp = s.get('http://localhost:8080/v1/config')
env = resp.json()

for key in env['featuresMap']:
    print(key)
```

## Run a feature flag rule

The Decide [endpoint](https://library.optimizely.com/docs/api/agent/v1/index.html#operation/decide) buckets a user into a feature flag variation (choosing between multiple enabled variations or one disabled variation) as part of a flag rule. Flag rules let you:
- experiment using A/B tests
- roll out feature flags progressively to a selected audience using targeted feature flag deliveries. 

To run a flag rule, use

```python
# decide 1 flag. 
params = { "keys": "my-feature-flag" }
payload = {
    "userId": "test-user",
    "userAttributes": {
        "attr1": "sample-attribute-1",
        "attr2": "sample-attribute-2"
    }
}

resp = s.post(url = 'http://localhost:8080/v1/decide', params=params, json=payload)

print(resp.json())


# multiple (bulk) feature flag decisions for specified flags.
# To decide ALL flags, simply omit keys params
payload = { "userId": "test-user" }
params = {"keys":"flag_1", "keys":"flag_2"}
resp2 = s.post(url = 'http://localhost:8080/v1/decide', params=params, json=payload)
print(json.dumps(resp.json(), indent=4, sort_keys=True))
```
The decide API is a POST to signal to the caller that there are side-effects. Namely, the decision results in a "decision" event sent to Optimizely analytics for the purpose of analyzing A/B test results. A decision event will NOT be sent by default if the flag is simply part of a delivery.
