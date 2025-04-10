#!/usr/bin/python
# example: python basic.py <SDK-Key>
# This basic example shows how to make individual decision requests
# with decide api

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

for key in env['featuresMap']:
    params = {"keys": key}
    resp = s.post(url='http://localhost:8080/v1/decide',
                  params=params, json=payload)
    print("Flag key: {}".format(key))
    print(json.dumps(resp.json(), indent=4, sort_keys=True))
