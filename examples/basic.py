#!/usr/bin/python
# example: python basic.py <SDK-Key>

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
for key in env['featuresMap']:
    params = {"featureKey": key}
    resp = s.post(url = 'http://localhost:8080/v1/activate', params=params, json=payload)
    print("Feature Key: {}".format(key))
    print(json.dumps(resp.json()[0], indent=4, sort_keys=True))

for key in env['experimentsMap']:
    params = {"experimentKey": key}
    resp = s.post(url = 'http://localhost:8080/v1/activate', params=params, json=payload)
    print("Experiment Key: {}".format(key))
    print(json.dumps(resp.json()[0], indent=4, sort_keys=True))

