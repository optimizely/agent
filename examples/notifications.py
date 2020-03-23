#!/usr/bin/python
# example: python notifications.py <SDK-Key>
# This example shows how to use the notifications endpoint to stream notifications using Server-Sent Events

import json
import requests
import sys
import threading

from sseclient import SSEClient

def print_notifications(sdk_key):
    # SSEClient (https://pypi.org/project/sseclient/) streams Server-Sent Events from the notifications endpoint
    # (/v1/notifications/event-stream), and we print each one
    messages = SSEClient('http://localhost:8080/v1/notifications/event-stream', headers={
        'X-Optimizely-Sdk-Key': sdk_key,
    })
    for msg in messages:
        print("Notification: {}".format(msg))


if len(sys.argv) < 2:
    sys.exit('Requires one argument: <SDK-Key>')

sdk_key = sys.argv[1]
thread = threading.Thread(target=print_notifications, args=(sdk_key,))
thread.start()

s = requests.Session()
s.headers.update({'X-Optimizely-SDK-Key': sdk_key})

resp = s.get('http://localhost:8080/v1/config')
env = resp.json()

payload = {
    "userId": "test-user",
    "userAttributes": {
        "attr1": "sample-attribute-1",
        "attr2": "sample-attribute-2"
    }
}

# Activate features and experiments - these will trigger notifications that will be sent to notification event stream
# subscribers

for key in env['featuresMap']:
    params = {"featureKey": key}
    s.post(url = 'http://localhost:8080/v1/activate', params=params, json=payload)

for key in env['experimentsMap']:
    params = {"experimentKey": key}
    s.post(url = 'http://localhost:8080/v1/activate', params=params, json=payload)

print("Send your own requests to trigger more notifications!")
thread.join()
