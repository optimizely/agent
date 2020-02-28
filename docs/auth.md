# Authorization Guide

## Overview

Optimizely Agent supports authorization workflows based on OAuth and JWT standards, allowing you to protect access to its API and Admin interfaces.

There are three modes of operation:

### Issuer & Validator
Access tokens are issued by Agent itself, using a [Client Credentials grant](https://www.oauth.com/oauth2-servers/access-tokens/client-credentials/). Access tokens are signed and validated using the HS256 algorithm with a signing secret provided in configuration.

### Validator-only
Agent validates access tokens that were issued elsewhere. Access tokens are validated with public keys fetched from a [JWKS](https://tools.ietf.org/html/rfc7517) URL provided in configuration.

### No authorization (default)
The interface is publicly available.

## Configuration
- The API and Admin interfaces are each independently configured to run in one of the above-mentioned modes of operation.
- Authorization configuration is applied under the `auth` key
- Each mode of operation has its own set of configuration properties, described below.

### Issuer & Validator
The configuration properties pertaining to Issuer & Validator mode are listed below:

|Property Name|Description|
|---|---|
|ttl|Time-to-live of access tokens issued|
|hmacSecrets|Secret used to sign issued access tokens, using the HMAC SHA256 algorithm|
|clients|Array of client credentials, any of which can be exchanged for an access token. Each object in the array should have `"id"` and `"secret"` string properties.|

### Validator-only
The configuration properties pertaining to Validator-only mode are listed below:

|Property Name|Description|
|---|---|
|jwksURL|URL from which public keys should be fetched for token validation|

### No authorization (default)
The API & Admin interfaces run with no authorization when no `auth` configuration is given.

### Example configuration file (yaml)
```yaml
# In this example, the API interface is configured in Validator-only mode, and the admin
# interface is configured in Issuer & Validation mode.

api:
    # Validator-only 
    auth:
        # Signing keys will be fetched from this url and used when validating access tokens
        jwksURL: https://YOUR_DOMAIN/.well-known/jwks.json
admin:
    # Issuer & Validator
    auth:
        # Access tokens will expire after 30 minutes
        ttl: 30m
         # Access tokens will be signed & validated using this secret
        hmacSecrets:
            - QPtUGP/RqaXRltZf1QE1KxlF2Iuo09J0buZ3UNKeIr0
        # Either of these two id/secret pairs can be exchanged for access tokens
        clients:
            - id: agentConsumer1
            secret: XgZTeTvWaZ6fLiey6EBSOxJ2QFdd6dIiUcZGDIIJ+IY 
            - id: agentConsumer2
            secret: ssz0EEViKIinkFXxzqncKxz+6VygEc2d2rKf+la5rXM 
```

## Secret Rotation (Issuer & Validator mode)
To support secret rotation, both `hmacSecrets` and `clients` are arrays. In `hmacSecrets`, the first array item will be
used to sign issued tokens, but tokens signed with any of the array items will be considered valid.
