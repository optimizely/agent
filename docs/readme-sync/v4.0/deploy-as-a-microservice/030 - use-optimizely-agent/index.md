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

Optimizely Agent provides [APIs](https://library.optimizely.com/docs/api/agent/v1/index.html) that enable running feature flag rules, such as A/B tests and targeted flag deliveries. Agent provides equivalent functionality to all our SDKs. At its core is the [Optimizely Go SDK](doc:go-sdk). 

### Running feature flag rules


The Decide [endpoint](https://library.optimizely.com/docs/api/agent/v1/index.html#operation/decide) buckets a user into a feature flag variation (choosing between multiple enabled or one disabled variation) as part of a flag rule. Flag rules include A/B tests and targeted feature flag deliveries. To run a flag rule, use:

`POST /v1/decide?keys={flagKey}`

In the request `application/json` body, include the `userID` and any `decideOptions`. The full request looks like this:

```curl
curl --request POST 'http://localhost:8080/v1/decide' \
--header 'Content-Type: application/json' \
--header 'X-Optimizely-SDK-Key: <YOUR_SDK_KEY>' \
--header 'Content-Type: application/json' \
--data-raw '{
"userId": "test-user"
"decideOptions": [
   "INCLUDE_REASONS"
]
}'
```

TODO: please review above and below request/response examples, I didn't test them, just looked at the Pull request for decide and made some guesses!! -FE



This returns an array of OptimizelyDecision objects that contains all the information you need to run your flag rule, such as:

- the decision to bucket this user into an enabled or disabled feature flag variation. 
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
	"enabled": false,
    "variationKey": "control_variation"
	"variables": {
		"my-var-1": "cust-val-default-1",
		"my-var-2": "cust-va1-default-2"
	},
    "reasons": []
}
```

The response is determined by the [A/B tests](https://docs.developers.optimizely.com/full-stack/v4.0/docs/run-a-b-tests) and [deliveries](https://docs.developers.optimizely.com/full-stack/v4.0/docs/run-flag-deliveries) defined for the supplied feature key, following the same rules as any Full Stack SDK. 

Note: If the user is bucketed into an A/B test, this endpoint dispatches a decision event.

### Authentication


To authenticate,  [pass your SDK key](https://docs.developers.optimizely.com/full-stack/docs/evaluate-rest-apis#section-start-an-http-session) as a header named ```X-Optimizely-SDK-Key``` in your API calls to Optimizely Agent. You can find your SDK key in app.optimizely.com under Settings > Environments > SDK Key. Remember you have a different SDK key for each environment. 

### Get All Decisions
- To get all feature flag decisions for a visitor in a single request, omit the feature flag parameter:
  `POST /v1/decide`
- To get decisions for multiple keys, specify multiple keys parameters, for example:
  `keys=flag_key_1&keys=flag_key_2`
  
  TODO: is above example right?

- To receive only the enabled feature flags for a visitor use a decide option in the `application/json` request body: 

```curl
--header 'Content-Type: application/json' \

--data-raw '{
"userId": "test-user"

"decideOptions": [
   "ENABLED_FLAGS_ONLY"
]

}'
```



### Tracking conversions

To track events, use the same  [tracking endpoint](https://library.optimizely.com/docs/api/agent/v1/index.html#operation/trackEvent) you use in the [SDKs' track API](doc:track-javascript):

`POST /v1/track?eventKey={eventKey}`

There is no response body for successful conversion event requests.

### API reference 

 For more  details on Optimizely Agent’s APIs, see the [complete API Reference](https://library.optimizely.com/docs/api/agent/v1/index.html).
