#!/usr/bin/python
# example: python auth.py <SDK-Key>

import json
import requests
import sys

# This example demonstrates obtaining an access token and using it to request the
# current Optimizely Config from the API interface.

# Before running, add the following to your Agent configuration file (default:
# config.yaml) under the "api" section:

# auth:
#   ttl: 30m
#   hmacSecrets:
#     - "abcd"
#   clients:
#     - id: clientid1
#       secret: clientsecret1

# With this configuration, an access token can be requested from the
# /oauth/apitoken endpoint using the configured id and secret. 

if len(sys.argv) < 2:
    sys.exit('Requires one argument: <SDK-Key>')

sdk_key = sys.argv[1]
client_id = sys.argv[2]
client_secret = sys.argv[3]

s = requests.Session()
s.headers.update({'X-Optimizely-SDK-Key': sdk_key})

resp = s.get('http://localhost:8080/v1/config')
print('first config request, not including access token: response status = {}'.format(resp.status_code))

resp = s.post('http://localhost:8080/oauth/api/token', data=json.dumps({
  'grant_type': 'client_credentials',
  'client_id':  client_id,
  'client_secret': client_secret,
}))
resp_dict = resp.json()
print('access token response: ')
print(json.dumps(resp_dict, indent=4, sort_keys=True))

s.headers.update({'Authorization': 'Bearer {}'.format(resp_dict['access_token'])})

resp = s.get('http://localhost:8080/v1/config')
print('config response after passing access token: ')
print(json.dumps(resp.json(), indent=4, sort_keys=True))
