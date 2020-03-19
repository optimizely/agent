#!/usr/bin/python
# example: python notifications.py <SDK-Key>
# This example shows how to attach and consume notifications from local running instance of the agent.

import sys
from sseclient import SSEClient
from six.moves.http_client import IncompleteRead

if len(sys.argv) < 2:
    sys.exit('Requires one argument: <SDK-Key>')

sdk_key = sys.argv[1]

def get_messages():
    print('here')
    messages = SSEClient('http://localhost:8080/v1/notifications/event-stream', timeout=100000, chunk_size=128,
                         headers={'X-Optimizely-SDK-Key': sdk_key})
    try:
        for msg in messages:
            print(msg)
    except IncompleteRead as e:
        print(e.partial)

while True:
    try:
        get_messages()
    except Exception as e:
        print('keep going')
