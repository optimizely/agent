#!/usr/bin/python
# example: python auth.py <SDK-Key>

import json
import requests
import sys

# This example demonstrates interacting with Agent running in Issuer & Validator mode.
# We obtain an access token and use it to request the current Optimizely Config
# from the API interface.
#
# Before running, add the following to your Agent configuration file (default:
# config.yaml) under the "api" section:
#
# auth:
#   ttl: 30m
#   clients:
#     - id: clientid1
#       secretHash: JDJhJDEyJFlkTWRIM1dEU3U3ZDNMTE1ZMGJoTU95a3lWZEUzaWRLWS5GWHpUSU05NHNlMTdnR09pdFJ1
#
# Then, set the following environment variable, which is the signing secret for access tokens Agent will issue:

# export OPTIMIZELY_API_AUTH_HMACSECRETS=llmO3xTUx+6TIfUU6eXmH/1Fh44ioL0h87G1iSrd5Gg
#
# With this configuration, an access token can be requested from the
# /oauth/apitoken endpoint using the configured client credentials.

# For more information, see docs/auth.md

if len(sys.argv) < 2:
    sys.exit('Requires one argument: <SDK-Key>')

sdk_key = sys.argv[1]
client_id = "clientid1"
client_secret = "bvz7VWyxgLUySOZQpzSblsy3/8570JjIPKWO06SjliA="

s = requests.Session()
s.headers.update({'X-Optimizely-SDK-Key': sdk_key})

resp = s.get('http://localhost:8080/v1/config')
print('first config request, not including access token: response status = {}'.format(resp.status_code))

resp = s.post('http://localhost:8080/oauth/token', data=json.dumps({
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
