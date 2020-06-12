---
title: "Evaluate REST APIs"
slug: "evaluate-rest-apis"
hidden: false
metadata: 
  title: "Evaluate REST APIs - Optimizely Full Stack"
createdAt: "2020-02-21T17:44:53.019Z"
updatedAt: "2020-04-13T23:02:34.056Z"
---
Below is an example demonstrating the APIs capabilities. For brevity, we've chosen to illustrate the API usage with Python. Note that the API documentation is defined via an OpenAPI (Swagger) spec and can be viewed [here](https://optimizely.github.io/docs/api/agent/).

## Start an http session
Each request made into Optimizely Agent is in the context of an Optimizely SDK Key. SDK Keys map API requests to a specific Optimizely Project and Environment. We can setup a global request header by using the `requests.Session` object.


[block:code]
{
  "codes": [
    {
      "code": "import requests\n\ns = requests.Session()\ns.headers.update({'X-Optimizely-SDK-Key': '<<YOUR-SDK-KEY>>'})",
      "language": "python"
    }
  ]
}
[/block]
The following examples will assume this session is being maintained.

## Get current environment configuration
The `/v1/config` endpoint returns a manifest of the current working environment.

[block:code]
{
  "codes": [
    {
      "code": "resp = s.get('http://localhost:8080/v1/config')\nenv = resp.json()\n\nfor key in env['featuresMap']:\n    print(key)",
      "language": "python"
    }
  ]
}
[/block]
## Activate Feature
The `/v1/activate?featureKey={key}` endpoint activates the feature for a given user. In Optimizely, activation is in the context of a given user to make the relative bucketing decision. In this case we'll provide a `userId` via the request body. The `userId` will be used to determine how the feature will be evaluated. Features can either be part of a Feature Test in which variations of feature variables are being measured against one another or a feature rollout, which progressively make the feature available to the selected audience.

From an API standpoint the presence of a Feature Test or Rollout is abstracted away from the response and only the resulting variation or enabled feature is returned.


[block:code]
{
  "codes": [
    {
      "code": "# single feature activate\nparams = { \"featureKey\": \"my-feature\" }\npayload = { \"userId\": \"test-user\" }\nresp = s.post(url = 'http://localhost:8080/v1/activate', params=params, json=payload)\n\nprint(resp.json())\n\n\n# multiple (bulk) feature activate\nparams = {\n    \"featureKey\": [key for key in env['featuresMap']],\n    \"experimentKey\": [key for key in env['experimentsMap']]\n}\nresp2 = s.post(url = 'http://localhost:8080/v1/activate', params=params, json=payload)\nprint(json.dumps(resp.json(), indent=4, sort_keys=True))",
      "language": "python"
    }
  ]
}
[/block]
The activate API is a POST to signal to the caller that there are side-effects. Namely, activation results in a "decision" event sent to Optimizely analytics for the purpose of analyzing Feature Test results. A "decision" will NOT be sent if the feature is simply part of a rollout.