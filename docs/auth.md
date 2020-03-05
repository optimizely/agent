# Authorization Guide

## Overview

Optimizely Agent supports authorization workflows based on OAuth and JWT standards, allowing you to protect access to its API and Admin interfaces.

There are three modes of operation:

### Issuer & Validator
Access tokens are issued by Agent itself, using a [Client Credentials grant](https://www.oauth.com/oauth2-servers/access-tokens/client-credentials/). Access tokens are signed and validated using the HS256 algorithm with a signing secret provided in configuration. Clients request access tokens by sending a `POST` request to `/oauth/token` on the port of the desired interface (by default, `8080` for the API interface, and `8088` for the Admin interface), including a client ID and secret in the request.


Issuer & Validator mode is useful if you want to implement authorization, and you are not already running an authorization server that can issue JWTs.

### Validator-only
Agent validates access tokens that were issued elsewhere. Access tokens are validated with public keys fetched from a [JWKS](https://tools.ietf.org/html/rfc7517) URL provided in configuration.

Validator-only mode is useful if you want to plug directly into an existing JWT-based workflow already being used in your system or organization.

### No authorization (default)
The interface is publicly available.

## Configuration
- The API and Admin interfaces are each independently configured to run in one of the above-mentioned modes of operation.
- Authorization configuration is located under the `auth` key
- Each mode of operation has its own set of configuration properties, described below.

### Issuer & Validator
The configuration properties pertaining to Issuer & Validator mode are listed below:

|Property Name|Environment Variable|Description|
|---|---|---|
|ttl|TTL|Time-to-live of access tokens issued|
|hmacSecrets|HMACSECRETS|Array of secrets used to sign & validate access tokens, using the HMAC SHA256 algorithm. The first value in the array is used to sign issued access tokens. Access tokens signed with any value in the array are considered valid.|
|clients|N/A|Array of `id` and `secretHash` pairs, used for access token issuance. Clients provide ID and secret in their requests to `/oauth/token`. Agent validates the request credentials by checking for an exact match of ID, and checking that the BCrypt hash of the request secret matches the `secretHash` from configuration. The `secretHash` in configuration is expected as a base64-format string.

To make setup easier, Agent provides a command-line tool that can generate base64-encoded 32-byte random values, and their associated base64-encoded BCrypt hashes:
```shell script
// From the Agent root directory
> make generate_secret
Client Secret: i3SrdrCy/wEGqggv9OI4FgIsdHHNpOacrmIMJ6SFIkE=
Client Secret's hash: JDJhJDEyJERGNzhjRXVTNTdOQUZ3cndxTkZ6Li5XQURlazU2R21YeFZjb1pWSkN5eGZ1SXM4VXRLb0ZD
```

Use the hash value to configure Agent, and pass the secret value as `client_secret` when making acces token requests to `/oauth/token`. For details of the access token issuance endpoint, see the OpenAPI spec file.

### Validator-only
The configuration properties pertaining to Validator-only mode are listed below:

|Property Name|Environment Variable|Description|
|---|---|---|
|jwksURL|JWKSURL|URL from which public keys should be fetched for token validation|

### No authorization (default)
The API & Admin interfaces run with no authorization when no `auth` configuration is given.

### Configuration examples
Optimizely Agent uses the [Viper](https://github.com/spf13/viper) library for configuration, which allows setting values via environment variables, flags, and YAML configuration files.
#### Issuer & Validator
_*WARNING*_: For security, we advise that you configure `hmacSecrets` with either an environment variable or a flag, and NOT through a config file.

In the below example, the Admin interface is configured in Issuer & Validator mode, with `hmacSecrets` provided via environment variable, and other values provided via YAML config file.

```shell script
// Comma-separated value, to set multiple hmacSecrets.
// Access tokens are signed with the first value.
// Access tokens are valid when they are signed with either of these values.
export OPTIMIZELY_ADMIN_HMACSECRETS=QPtUGP/RqaXRltZf1QE1KxlF2Iuo09J0buZ3UNKeIr0,bkZAqSsZuM5NSnwEyO9Pzb6F8gGNu1BBuX/SpPaMeyM
```

```yaml
# config.yaml
admin:
    auth:
        # Access tokens will expire after 30 minutes
        ttl: 30m
        clients:
            # Either of these two id/secret pairs can be exchanged for access tokens
            - id: agentConsumer1
            secretHash: XgZTeTvWaZ6fLiey6EBSOxJ2QFdd6dIiUcZGDIIJ+IY 
            - id: agentConsumer2
            secretHash: ssz0EEViKIinkFXxzqncKxz+6VygEc2d2rKf+la5rXM 
```
#### Validator-only
```yaml
# In this example, the API interface is configured in Validator-only mode
api:
    auth:
        # Signing keys will be fetched from this url and used when validating access tokens
        jwksURL: https://YOUR_DOMAIN/.well-known/jwks.json
```

## Secret Rotation (Issuer & Validator mode)
To support secret rotation, both `hmacSecrets` and `clients` support setting multiple values. In `hmacSecrets`, the first value will be
used to sign issued tokens, but tokens signed with any of the values will be considered valid.

## Example (Python)
For example requests demonstrating the Issuer & Validator mode, see [here](../examples/auth.py).
