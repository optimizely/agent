#!/usr/bin/python

import json
import requests
import sys

sdk_key = sys.argv[1]
event_key = sys.argv[2]

s = requests.Session()
s.headers.update({'X-Optimizely-SDK-Key': sdk_key})

params = { "eventKey": event_key }
payload = { "userId": "test-user" }

resp = s.post('http://localhost:8080/v1/track', params=params, json=payload)
print(resp)

