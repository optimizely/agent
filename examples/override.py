#!/usr/bin/python
# example: python override.py <SDK-Key> <Experiment-Key> <Variation-Key>

import json
import requests
import sys

if len(sys.argv) < 4:
    sys.exit('Requires three arguments: <SDK-Key> <Experiment-Key> <Variation-Key>')

sdk_key = sys.argv[1]
exp_key = sys.argv[2]
var_key = sys.argv[3]

s = requests.Session()
s.headers.update({'X-Optimizely-SDK-Key': sdk_key})

payload = {
    "userId": "test-user",
    "experimentKey": exp_key,
    "variationKey": var_key
}

resp = s.post('http://localhost:8080/v1/override', json=payload)
print(json.dumps(resp.json(), indent=4, sort_keys=True))
