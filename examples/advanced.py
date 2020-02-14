#!/usr/bin/python
# example: python advanced.py <SDK-Key>

import json
import requests
import sys

if len(sys.argv) < 2:
    sys.exit('Requires one argument: <SDK-Key>')

sdk_key = sys.argv[1]

s = requests.Session()
s.headers.update({'X-Optimizely-SDK-Key': sdk_key})

resp = s.get('http://localhost:8080/v1/config')
env = resp.json()

payload = { "userId": "test-user" }

params = {
            "featureKey": [key for key in env['featuresMap']],
            "experimentKey": [key for key in env['experimentsMap']]
        }

resp = s.post(url = 'http://localhost:8080/v1/activate', params=params, json=payload)
print(json.dumps(resp.json(), indent=4, sort_keys=True))

