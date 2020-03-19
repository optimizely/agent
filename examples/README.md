## Optimizely Agent Examples

This folder provides a set of sample scripts to illustrate some of the main API features.

### Installation

The included examples were written in python and assumes python 3.7+. If using a python3 venv [virtual environment](https://packaging.python.org/guides/installing-using-pip-and-virtual-environments/)
you can install all of the dependencies in the requirement.txt:
```bash
pip install -r requirements.txt
```

### Activate

The `/activate` endpoint returns a decision of the requested experiment or feature for a given user context.
For single decisions please refer to [basic.py](./basic.py) which demonstrates how to iterate through the configuration for
a given SDK project and make individual activation requests.

Example usage:
```bash
python basic.py <SDK-Key>
```

The `/activate` endpoint also supports batching requests by supplying multiple experiment or feature keys in 
single `/activate` call. Please refer to [advanced.py](./advanced.py) for an example on how to fetch multiple decisions. This
is useful to reduce the number of outbound API calls made from your service to Optimizely Agent.

Example usage:
```bash
python advanced.py <SDK-Key>
```

### Track

The `/track` endpoint is used to send conversion events to the Optimizely analytics backend.
The example in [track.py](./track.py) provides an example of calling the /track api with a set of event tags
in the payload. 

Example usage:
```bash
python track.py <SDK-Key> <Event-Key>
```

### Override

The `/override` endpoint provides the ability to force a particular experiment decision for a given user.
This endpoint is disabled by default and should not be used in a production environment and is recommended
for development and QA testing. An example of how to set an override can be found in `override.py`.

Example usage:
```bash
python override.py <SDK-Key> <Experiment-Key> <Variation-Key>
```

### Auth

Optimizely Agent supports a [client_credentials](https://www.oauth.com/oauth2-servers/access-tokens/client-credentials/)
grant type for applications to requests access tokens for a set of SDK resources. The `auth.py` script demonstrates
how to request an authorization token and use it in subsequent API requests. Please refer to [auth.md](../docs/auth.md)
for a complete overview of the authentication modes supported by Agent.

Example usage:
```bash
python auth.py <Your SDK Key> clientid1 0bfLVX9U3Lpr6Qe4X3DSSIWNqEkEQ4bkX1WZ5Km6spM=
```