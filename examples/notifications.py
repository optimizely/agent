from sseclient import SSEClient
from six.moves.http_client import IncompleteRead

def get_messages():
    requests_kwargs = {'headers': {}}
    requests_kwargs['headers']['X-Optimizely-SDK-Key'] = 'FeEqVPw2ZKLcxjHX5A732L'
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