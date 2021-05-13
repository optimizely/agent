#!/usr/bin/python
# example: python advanced_activate.py <SDK-Key>
# This advanced example shows how to make batched activation requests.

import json
import requests
import sys

if len(sys.argv) < 2:
    sys.exit('Requires one argument: <SDK-Key>')

sdk_key = sys.argv[1]

s = requests.Session()
s.headers.update({'X-Optimizely-SDK-Key': sdk_key})

# Making a request to /config to generically pull the set of features and experiments.
# In production, making this initial request to build the set of keys would not be recommended.
# Instead the keys would already be known by the application, or we'd use the type= parameter illustrated below.
resp = s.get('http://localhost:8080/v1/config')
env = resp.json()

payload = {
    "userId": "test-user",
    "userAttributes": {
        "attr1": "sample-attribute-1",
        "attr2": "sample-attribute-2"
    }
}

# /activate accepts a list of feature and/or experiment keys
params = {
    "featureKey": [key for key in env['featuresMap']],
    "experimentKey": [key for key in env['experimentsMap']]
}
resp = s.post(url = 'http://localhost:8080/v1/activate', params=params, json=payload)
print(json.dumps(resp.json(), indent=4, sort_keys=True))

# Alternatively /activate can be passed a type of either "feature" or "experiment"
params = {"type": ["experiment", "feature"]}
resp = s.post(url = 'http://localhost:8080/v1/activate', params=params, json=payload)
print(json.dumps(resp.json(), indent=4, sort_keys=True))
