#!/usr/bin/python

import json
import requests
import sys

# This example demonstrates interacting with Agent running in Issuer & Validator mode.
# We obtain an access token and use it to request the current Optimizely Config
# from the API interface.

# Fist, we need a secret value to sign access tokens.
# You can use the generate_secret tool included with Agent to generate this:

# > make generate_secret
# Secret: CvzvkWm3V1D9RBxPWEjC+ud9zvwcOvnnLkWaIkzDGyA=

# You can ignore the second line that says "Secret's hash".

# Then, set an environment variable to make this secret available to Agent:
# > export OPTIMIZELY_API_AUTH_HMACSECRETS=CvzvkWm3V1D9RBxPWEjC+ud9zvwcOvnnLkWaIkzDGyA=

# Next, we need client credentials (ID & secret), and the BCrypt hash of our secret
# Again, you can use the generate_secret tool included with Agent to generate these:
#
# > make generate_secret
# Secret: 0bfLVX9U3Lpr6Qe4X3DSSIWNqEkEQ4bkX1WZ5Km6spM=
# Secret's hash: JDJhJDEyJEdkSHpicHpRODBqOC9FQzRneGIyNXU0ZFVPMFNKcUhkdTRUQXRzWUJOdjRzRmcuVGdFUTUu
#
# Take the hash, and add it to your agent configuration file (default: config.yaml) under the "api" section,
# along with your desired client ID and SDK key:
#
# auth:
#   ttl: 30m
#   clients:
#     - id: clientid1
#       secretHash: JDJhJDEyJEdkSHpicHpRODBqOC9FQzRneGIyNXU0ZFVPMFNKcUhkdTRUQXRzWUJOdjRzRmcuVGdFUTUu
#       sdkKeys:
#           - <Your SDK Key>

#
# Start Agent with the API interface running on the default port (8080).
# Then, finally, run the example, passing your SDK key, client ID and secret:
# > python auth.py <Your SDK Key> clientid1 0bfLVX9U3Lpr6Qe4X3DSSIWNqEkEQ4bkX1WZ5Km6spM=
#
# For more information, see docs/auth.md

if len(sys.argv) < 4:
    sys.exit('Requires three arguments: <SDK-Key> <Client ID> <Client Secret>')

sdk_key = sys.argv[1]
client_id = sys.argv[2]
client_secret = sys.argv[3]

s = requests.Session()
s.headers.update({'X-Optimizely-SDK-Key': sdk_key})

resp = s.get('http://localhost:8080/v1/config')
print('first config request, not including access token: response status = {}'.format(resp.status_code))

resp = s.post('http://localhost:8080/oauth/token', data={
    'grant_type': 'client_credentials',
    'client_id':  client_id,
    'client_secret': client_secret,
})
resp_dict = resp.json()
print('access token response: ')
print(json.dumps(resp_dict, indent=4, sort_keys=True))

s.headers.update({'Authorization': 'Bearer {}'.format(resp_dict['access_token'])})

resp = s.get('http://localhost:8080/v1/config')
print('config response after passing access token: ')
print(json.dumps(resp.json(), indent=4, sort_keys=True))
