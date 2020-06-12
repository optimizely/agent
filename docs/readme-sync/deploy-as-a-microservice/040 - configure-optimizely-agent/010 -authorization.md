---
title: "Authorization"
slug: "authorization"
hidden: false
metadata: 
  title: "Agent Authorization - Optimizely Full Stack"
createdAt: "2020-03-11T20:58:11.777Z"
updatedAt: "2020-03-31T19:44:52.119Z"
---
Optimizely Agent supports authorization workflows based on OAuth and JWT standards, allowing you to protect access to its API and Admin interfaces.

There are three modes of operation:

##1. Issuer & Validator
Access tokens are issued by Agent itself, using a [Client Credentials grant](https://www.oauth.com/oauth2-servers/access-tokens/client-credentials/). Access tokens are signed and validated using the HS256 algorithm with a signing secret provided in configuration. Clients request access tokens by sending a `POST` request to `/oauth/token` on the port of the desired interface (by default, `8080` for the API interface, and `8088` for the Admin interface), including a client ID and secret in the request.


Issuer & Validator mode is useful if you want to implement authorization, and you are not already running an authorization server that can issue JWTs.

##2. Validator-only
Agent validates access tokens that were issued elsewhere. Access tokens are validated with public keys fetched from a [JWKS](https://tools.ietf.org/html/rfc7517) URL provided in configuration.

Validator-only mode is useful if you want to plug directly into an existing JWT-based workflow already being used in your system or organization.

##3. No authorization (default)
The interface is publicly available.

# Configuration
- The API and Admin interfaces are each independently configured to run in one of the above-mentioned modes of operation.
- Authorization configuration is located under the `auth` key
- Each mode of operation has its own set of configuration properties, described below.

## Issuer & Validator
The configuration properties pertaining to Issuer & Validator mode are listed below:

|Property Name|Environment Variable|Description|
|---|---|---|
|ttl|TTL|Time-to-live of access tokens issued|
|hmacSecrets|HMACSECRETS|Array of secrets used to sign & validate access tokens, using the HMAC SHA256 algorithm. Values must be base64-format strings. The first value in the array is used to sign issued access tokens. Access tokens signed with any value in the array are considered valid.|
|clients|N/A|Array of objects, used for token issuance, consisting of `id`, `secretHash`, and `sdkKeys`. Clients provide ID and secret in their requests to `/oauth/token`. Agent validates the request credentials by checking for an exact match of ID, checking that the BCrypt hash of the request secret matches the `secretHash` from configuration, and that the SDK key provided in the `X-Optimizely-Sdk-Key` request header exists in the `sdkKeys` from configuration. `secretHash` values must be base64-format strings.|

To make setup easier, Agent provides a command-line tool that can generate base64-encoded 32-byte random values, and their associated base64-encoded BCrypt hashes:

[block:code]
{
  "codes": [
    {
      "code": "// From the Agent root directory\n> make generate_secret\nClient Secret: i3SrdrCy/wEGqggv9OI4FgIsdHHNpOacrmIMJ6SFIkE=\nClient Secret's hash: JDJhJDEyJERGNzhjRXVTNTdOQUZ3cndxTkZ6Li5XQURlazU2R21YeFZjb1pWSkN5eGZ1SXM4VXRLb0ZD",
      "language": "shell"
    }
  ]
}
[/block]
Use the hash value to configure Agent, and pass the secret value as `client_secret` when making access token requests to `/oauth/token`. For details of the access token issuance endpoint, see the OpenAPI spec file.

## Validator-only
The configuration properties pertaining to Validator-only mode are listed below:

|Property Name|Environment Variable|Description|
|---|---|---|
|jwksURL|JWKSURL|URL from which public keys should be fetched for token validation|
|jwksUpdateInterval|JWKSUPDATEINTERVAL|Interval on which public keys should be re-fetched (example: `30m` for 30 minutes)|

## No authorization (default)
The API & Admin interfaces run with no authorization when no `auth` configuration is given.

## Configuration examples
Optimizely Agent uses the [Viper](https://github.com/spf13/viper) library for configuration, which allows setting values via environment variables, flags, and YAML configuration files.
### Issuer & Validator
_*WARNING*_: For security, we advise that you configure `hmacSecrets` with either an environment variable or a flag, and NOT through a config file.

In the below example, the Admin interface is configured in Issuer & Validator mode, with `hmacSecrets` provided via environment variable, and other values provided via YAML config file.
[block:code]
{
  "codes": [
    {
      "code": "// Comma-separated value, to set multiple hmacSecrets.\n// Access tokens are signed with the first value.\n// Access tokens are valid when they are signed with either of these values.\nexport OPTIMIZELY_ADMIN_HMACSECRETS=QPtUGP/RqaXRltZf1QE1KxlF2Iuo09J0buZ3UNKeIr0,bkZAqSsZuM5NSnwEyO9Pzb6F8gGNu1BBuX/SpPaMeyM",
      "language": "shell"
    }
  ]
}
[/block]

[block:code]
{
  "codes": [
    {
      "code": "admin:\n    auth:\n        # Access tokens will expire after 30 minutes\n        ttl: 30m\n        clients:\n            # Either of these two id/secret pairs can be exchanged for access tokens\n            - id: agentConsumer1\n            secretHash: XgZTeTvWaZ6fLiey6EBSOxJ2QFdd6dIiUcZGDIIJ+IY \n            sdkKeys:\n              # These credentials can be exchanged for tokens granting access to these two SDK keys\n              - abcd1234\n              - efgh5678\n            - id: agentConsumer2\n            secretHash: ssz0EEViKIinkFXxzqncKxz+6VygEc2d2rKf+la5rXM \n            sdkKeys:\n              # These credentials can be exchanged for tokens granting access only to this one SDK key\n              - ijkl9012",
      "language": "yaml",
      "name": "config.yaml"
    }
  ]
}
[/block]

### Validator-only
```yaml
# In this example, the API interface is configured in Validator-only mode
api:
    auth:
        # Signing keys will be fetched from this url and used when validating access tokens
        jwksURL: https://YOUR_DOMAIN/.well-known/jwks.json
        # Siging keys will be periodically fetched on this interval
        jwksUpdateInterval: 30m
```

# Secret Rotation (Issuer & Validator mode)
To support secret rotation, both `hmacSecrets` and `clients` support setting multiple values. In `hmacSecrets`, the first value will be
used to sign issued tokens, but tokens signed with any of the values will be considered valid.

# Example (Python)
Example requests demonstrating the Issuer & Validator mode:
[block:code]
{
  "codes": [
    {
      "code": "#!/usr/bin/python\n\nimport json\nimport requests\nimport sys\n\n# This example demonstrates interacting with Agent running in Issuer & Validator mode.\n# We obtain an access token and use it to request the current Optimizely Config\n# from the API interface.\n\n# Fist, we need a secret value to sign access tokens.\n# You can use the generate_secret tool included with Agent to generate this:\n\n# > make generate_secret\n# Client Secret: CvzvkWm3V1D9RBxPWEjC+ud9zvwcOvnnLkWaIkzDGyA=\n\n# You can ignore the second line that says \"Client Secret's hash\".\n\n# Then, set an environment variable to make this secret available to Agent:\n# > export OPTIMIZELY_API_AUTH_HMACSECRETS=CvzvkWm3V1D9RBxPWEjC+ud9zvwcOvnnLkWaIkzDGyA=\n\n# Next, we need client credentials (ID & secret), and the BCrypt hash of our secret\n# Again, you can use the generate_secret tool included with Agent to generate these:\n#\n# > make generate_secret\n# Client Secret: 0bfLVX9U3Lpr6Qe4X3DSSIWNqEkEQ4bkX1WZ5Km6spM=\n# Client Secret's hash: JDJhJDEyJEdkSHpicHpRODBqOC9FQzRneGIyNXU0ZFVPMFNKcUhkdTRUQXRzWUJOdjRzRmcuVGdFUTUu\n#\n# Take the hash, and add it to your agent configuration file (default: config.yaml) under the \"api\" section,\n# along with your desired client ID and SDK key:\n#\n# auth:\n#   ttl: 30m\n#   clients:\n#     - id: clientid1\n#       secretHash: JDJhJDEyJEdkSHpicHpRODBqOC9FQzRneGIyNXU0ZFVPMFNKcUhkdTRUQXRzWUJOdjRzRmcuVGdFUTUu\n#       sdkKeys:\n#           - <Your SDK Key>\n\n#\n# Start Agent with the API interface running on the default port (8080).\n# Then, finally, run the example, passing your SDK key, client ID and secret:\n# > python auth.py <Your SDK Key> clientid1 0bfLVX9U3Lpr6Qe4X3DSSIWNqEkEQ4bkX1WZ5Km6spM=\n#\n# For more information, see docs/auth.md\n\nif len(sys.argv) < 4:\n    sys.exit('Requires three arguments: <SDK-Key> <Client ID> <Client Secret>')\n\nsdk_key = sys.argv[1]\nclient_id = sys.argv[2]\nclient_secret = sys.argv[3]\n\ns = requests.Session()\ns.headers.update({'X-Optimizely-SDK-Key': sdk_key})\n\nresp = s.get('http://localhost:8080/v1/config')\nprint('first config request, not including access token: response status = {}'.format(resp.status_code))\n\nresp = s.post('http://localhost:8080/oauth/token', data={\n    'grant_type': 'client_credentials',\n    'client_id':  client_id,\n    'client_secret': client_secret,\n})\nresp_dict = resp.json()\nprint('access token response: ')\nprint(json.dumps(resp_dict, indent=4, sort_keys=True))\n\ns.headers.update({'Authorization': 'Bearer {}'.format(resp_dict['access_token'])})\n\nresp = s.get('http://localhost:8080/v1/config')\nprint('config response after passing access token: ')\nprint(json.dumps(resp.json(), indent=4, sort_keys=True))",
      "language": "python"
    }
  ]
}
[/block]