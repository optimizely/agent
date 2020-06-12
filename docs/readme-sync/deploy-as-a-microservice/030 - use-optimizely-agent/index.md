---
title: "Use Optimizely Agent"
slug: "use-optimizely-agent"
hidden: false
metadata: 
  title: "How to use Optimizely Agent - Optimizely Full Stack"
createdAt: "2020-02-21T17:44:28.054Z"
updatedAt: "2020-04-08T21:26:30.308Z"
---
Optimizely Agent provides [APIs](https://optimizely.github.io/docs/api/agent/) that enable experimentation and feature management. Agent provides equivalent functionality to all our SDKs. At its core is the [Optimizely Go SDK](doc:go-sdk). In some cases, however, we’ve updated our APIs to simplify key use cases.
[block:api-header]
{
  "title": "Manage features"
}
[/block]
 Optimizely Agent simplifies the core feature management of our [SDK APIs](doc:sdk-reference-guides).  It consolidates the following endpoints:

- [isFeatureEnabled](doc:is-feature-enabled-go)
- [getFeatureVariableBoolean](doc:get-feature-variable-go#section-boolean)
- [getFeatureVariableDouble](doc:get-feature-variable-go#section-double)
- [getFeatureVariableInteger](doc:get-feature-variable-go#section-integer)
- [getFeatureVariableString](doc:get-feature-variable-go#section-string) 

... into one, convenient endpoint:

`POST /v1/activate?featureKey={featureKey}`

This [endpoint](https://optimizely.github.io/docs/api/agent/#operation/activate) returns:

- the decision for this feature for this user
- any corresponding feature variable values. 

For example: 
[block:code]
{
  "codes": [
    {
      "code": "{\n\t\"featureKey\": \"feature-key-1\",\n\t\"enabled\": true,\n\t\"variables\": {\n\t\t\"my-var-1\": \"cust-val-1\",\n\t\t\"my-var-2\": \"cust-va1-2\"\n\t}\n}",
      "language": "json"
    }
  ]
}
[/block]
The response is determined by the [feature tests](doc:run-feature-tests) and [feature rollouts](doc:use-feature-flags) defined for the supplied feature key, following the same rules as any Full Stack SDK. 

Note: If the user is assigned to a feature test, this API will dispatch an impression.

[block:api-header]
{
  "title": "Authentication"
}
[/block]
To authenticate,  [pass your SDK key](https://docs.developers.optimizely.com/full-stack/docs/evaluate-rest-apis#section-start-an-http-session) as a header named ```X-Optimizely-SDK-Key``` in your API calls to Optimizely Agent. You can find your SDK key in app.optimizely.com under Settings > Environments > SDK Key. Remember you have a different SDK key for each environment. 
[block:api-header]
{
  "title": "Running A/B Tests"
}
[/block]
To activate an A/B test, use:

`POST /v1/activate?experimentKey={experimentKey}`

This dispatches an impression and return the user’s assigned variation:
[block:code]
{
  "codes": [
    {
      "code": "{\n  \"experimentKey\": \"experiment-key-1\"\n\t\"variationKey\": \"variation-key-1\"\n}",
      "language": "json"
    }
  ]
}
[/block]

[block:api-header]
{
  "title": "Tracking Conversions"
}
[/block]
To track events, use the same  [tracking endpoint](https://optimizely.github.io/docs/api/agent/#operation/trackEvent) you use in the [SDKs' track API](doc:track-javascript):

`POST /v1/track?eventKey={eventKey}`

There is no response body for successful conversion event requests.
[block:api-header]
{
  "title": "API Reference"
}
[/block]
 For more  details on Optimizely Agent’s APIs, see the [complete API Reference](https://optimizely.github.io/docs/api/agent/).