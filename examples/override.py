#!/usr/bin/python

import json
import requests
import sys

sdk_key = sys.argv[1]
exp_key = sys.argv[2]
var_key = sys.argv[3]

s = requests.Session()
s.headers.update({'X-Optimizely-SDK-Key': sdk_key})

payload = { "userId": "test-user", "experimentKey": exp_key, "variationKey": var_key }

resp = s.post('http://localhost:8080/v1/override', json=payload)
print(resp)

