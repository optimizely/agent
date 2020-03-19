#!/usr/bin/python
# example: python notifications.py <SDK-Key>
# This example shows how to attach and consume notifications from local running instance of the agent.

from sseclient import SSEClient
from six.moves.http_client import IncompleteRead

def get_messages():
    print('here')
    messages = SSEClient('http://localhost:8080/v1/notifications/event-stream', timeout=100000, chunk_size=128,
                         headers={'X-Optimizely-SDK-Key': 'FeEqVPw2ZKLcxjHX5A732L'})
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
