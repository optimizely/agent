#!/usr/bin/python
# example: python advanced.py <SDK-Key>
# This advanced example shows:
# 1. The result for a single key is returned as an OptimizelyDecision object.
# 2. The result for multiple keys is returned as an array of OptimizelyDecision objects.
# 3. When no flag key is provided, decision is made for all flag keys.

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
resp = s.get('http://localhost:8080/v1/config')
env = resp.json()

payload = {
    "userId": "test-user",
    "decideOptions": [
        "ENABLED_FLAGS_ONLY",
        "INCLUDE_REASONS"
    ],
    "userAttributes": {
        "attr1": "sample-attribute-1",
        "attr2": "sample-attribute-2"
    }
}

# The result for a single key is returned as an OptimizelyDecision object
key = [key for key in env['featuresMap']][0]
params = {"keys": key}
resp = s.post(url = 'http://localhost:8080/v1/decide', params=params, json=payload)
print("OptimizelyDecision object for flag key {}".format(key))
print(json.dumps(resp.json(), indent=4, sort_keys=True))

# The result for multiple keys is returned as an array of OptimizelyDecision objects
keys = [key for key in env['featuresMap']]
params = {"keys": keys}
resp = s.post(url = 'http://localhost:8080/v1/decide', params=params, json=payload)
print("Array of OptimizelyDecision objects")
print(json.dumps(resp.json(), indent=4, sort_keys=True))

# When no flag key is provided, decision is made for all flag keys.
params = {"keys": None}
resp = s.post(url = 'http://localhost:8080/v1/decide', params=params, json=payload)
print("Decision for all flag keys when flagh key is not provided.")
print(json.dumps(resp.json(), indent=4, sort_keys=True))
