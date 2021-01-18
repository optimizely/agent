---
title: "Use Optimizely Agent"
excerpt: ""
slug: "use-optimizely-agent"
hidden: false
metadata: 
  title: "How to use Optimizely Agent - Optimizely Full Stack"
createdAt: "2020-02-21T17:44:28.054Z"
updatedAt: "2020-04-08T21:26:30.308Z"
---

Optimizely Agent provides [APIs](https://library.optimizely.com/docs/api/agent/v1/index.html) that enable feature flag-based A/B tests and feature flag delivery. Agent provides equivalent functionality to all our SDKs. At its core is the [Optimizely Go SDK](doc:go-sdk). 

### Decide flags


The Decide [endpoint](https://library.optimizely.com/docs/api/agent/v1/index.html#operation/decide):

`POST /v1/decide?keys={flagKey}`

returns an array of OptimizelyDecision objects that contain information such as:

- the decision for this flag for this user
- any corresponding feature flag variable values. 

For example: 

```json
{
	"userContext": {
        "UserId": "test-user",
        "Attributes": {
            "logged-in": true,
            "location": "usa"
        }
    },
    "flagKey": "my-feature-flag",
    "ruleKey": "my-a-b-test",
	"enabled": true,
	"variables": {
		"my-var-1": "cust-val-1",
		"my-var-2": "cust-va1-2"
	},
    "reasons": [""]
}
```

The response is determined by the [A/B tests](https://docs.developers.optimizely.com/full-stack/v4.0/docs/run-a-b-tests) and [deliveries](https://docs.developers.optimizely.com/full-stack/v4.0/docs/run-flag-deliveries) defined for the supplied feature key, following the same rules as any Full Stack SDK. 

Note: If the user is assigned to an A/B test, this API will dispatch a decision event.

### Authentication


To authenticate,  [pass your SDK key](https://docs.developers.optimizely.com/full-stack/docs/evaluate-rest-apis#section-start-an-http-session) as a header named ```X-Optimizely-SDK-Key``` in your API calls to Optimizely Agent. You can find your SDK key in app.optimizely.com under Settings > Environments > SDK Key. Remember you have a different SDK key for each environment. 


### Running A/B tests


To activate an A/B test, use:

`POST /v1/activate?experimentKey={experimentKey}`

This dispatches an impression and return the user’s assigned variation:

`POST /v1/activate?experimentKey={experimentKey}`

This dispatches an impression and return the user’s assigned variation:
```json
{
  "experimentKey": "experiment-key-1",
  "variationKey": "variation-key-1"
}
```
### Get All Decisions
To get all Feature decisions for a visitor in a single request use:
`POST /v1/activate?type=feature`

To receive only the enabled features for a visitor use: 

`POST /v1/activate?type=feature&enabled=true`

To get all Experiment decisions for a visitor in a single request use:
`POST /v1/activate?type=experiment`



### Tracking conversions

To track events, use the same  [tracking endpoint](https://library.optimizely.com/docs/api/agent/v1/index.html#operation/trackEvent) you use in the [SDKs' track API](doc:track-javascript):

`POST /v1/track?eventKey={eventKey}`

There is no response body for successful conversion event requests.

### API reference 

 For more  details on Optimizely Agent’s APIs, see the [complete API Reference](https://library.optimizely.com/docs/api/agent/v1/index.html).
