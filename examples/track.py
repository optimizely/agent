#!/usr/bin/python

import json
import requests
import sys

if len(sys.argv) < 3:
    sys.exit('Requires two arguments: <SDK-Key> <Event-Key>')

sdk_key = sys.argv[1]
event_key = sys.argv[2]

s = requests.Session()
s.headers.update({'X-Optimizely-SDK-Key': sdk_key})

params = {"eventKey": event_key}
payload = {
    "userId": "test-user",
    "userAttributes": {
        "attr1": "sample-attribute-1",
        "attr2": "sample-attribute-2"
    },
    "eventTags": {
        "event-tag-1": "custom-tag-1",
        "event-tag-2": "custom-tag-2"
    }
}

resp = s.post('http://localhost:8080/v1/track', params=params, json=payload)
print(resp)

