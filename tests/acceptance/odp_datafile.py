odp_datafile = {
    "groups": [

    ],
    "environmentKey": "production",
    "rollouts": [
        {
            "experiments": [
                {
                    "status": "Running",
                    "audienceConditions": [

                    ],
                    "audienceIds": [

                    ],
                    "variations": [
                        {
                            "variables": [

                            ],
                            "id": "151743",
                            "key": "off",
                            "featureEnabled": False
                        }
                    ],
                    "forcedVariations": {

                    },
                    "key": "default-rollout-52207-23726430538",
                    "layerId": "rollout-52207-23726430538",
                    "trafficAllocation": [
                        {
                            "entityId": "151743",
                            "endOfRange": 10000
                        }
                    ],
                    "id": "default-rollout-52207-23726430538"
                }
            ],
            "id": "rollout-52207-23726430538"
        },
        {
            "experiments": [
                {
                    "status": "Running",
                    "audienceConditions": [

                    ],
                    "audienceIds": [

                    ],
                    "variations": [
                        {
                            "variables": [

                            ],
                            "id": "151797",
                            "key": "off",
                            "featureEnabled": False
                        }
                    ],
                    "forcedVariations": {

                    },
                    "key": "default-rollout-52231-23726430538",
                    "layerId": "rollout-52231-23726430538",
                    "trafficAllocation": [
                        {
                            "entityId": "151797",
                            "endOfRange": 10000
                        }
                    ],
                    "id": "default-rollout-52231-23726430538"
                }
            ],
            "id": "rollout-52231-23726430538"
        }
    ],
    "typedAudiences": [
        {
            "id": "23783030150",
            "conditions": [
                "and",
                [
                  "or",
                  [
                      "or",
                      {
                          "value": "mac",
                          "type": "custom_attribute",
                          "name": "laptop_os",
                          "match": "exact"
                      }
                  ],
                    [
                      "or",
                      {
                          "value": "atsbugbashsegmentdob",
                          "type": "third_party_dimension",
                          "name": "odp.audiences",
                          "match": "qualified"
                      }
                  ],
                    [
                      "or",
                      {
                          "value": "atsbugbashsegmentgender",
                          "type": "third_party_dimension",
                          "name": "odp.audiences",
                          "match": "qualified"
                      }
                  ],
                    [
                      "or",
                      {
                          "value": "atsbugbashsegmenthaspurchased",
                          "type": "third_party_dimension",
                          "name": "odp.audiences",
                          "match": "qualified"
                      }
                  ]
                ]
            ],
            "name": "Customer Information"
        }
    ],
    "projectId": "23743870473",
    "variables": [

    ],
    "featureFlags": [
        {
            "experimentIds": [
                "9300000235908"
            ],
            "rolloutId": "rollout-52207-23726430538",
            "variables": [

            ],
            "id": "52207",
            "key": "flag1"
        },
        {
            "experimentIds": [

            ],
            "rolloutId": "rollout-52231-23726430538",
            "variables": [

            ],
            "id": "52231",
            "key": "flag2"
        }
    ],
    "integrations": [
        {
            "publicKey": "ax6UV2223fD-jpOXID0BMg",
            "host": "https://api.zaius.com",
            "key": "odp",
            "pixelUrl": "https://jumbe.zaius.com",
        }
    ],
    "experiments": [
        {
            "status": "Running",
            "audienceConditions": [
                "or",
                "23783030150"
            ],
            "audienceIds": [
                "23783030150"
            ],
            "variations": [
                {
                    "variables": [

                    ],
                    "id": "151799",
                    "key": "variation_a",
                    "featureEnabled": True
                },
                {
                    "variables": [

                    ],
                    "id": "151800",
                    "key": "variation_b",
                    "featureEnabled": True
                }
            ],
            "forcedVariations": {

            },
            "key": "ab_experiment",
            "layerId": "9300000198165",
            "trafficAllocation": [
                {
                    "entityId": "151799",
                    "endOfRange": 5000
                },
                {
                    "entityId": "151800",
                    "endOfRange": 10000
                }
            ],
            "id": "9300000235908"
        }
    ],
    "version": "4",
    "audiences": [
        {
            "id": "23783030150",
            "conditions": "[\"or\", {\"match\": \"exact\", \"name\": \"$opt_dummy_attribute\", \"type\": \"custom_attribute\", \"value\": \"$opt_dummy_value\"}]",
            "name": "Customer Information"
        },
        {
            "conditions": "[\"or\", {\"match\": \"exact\", \"name\": \"$opt_dummy_attribute\", \"type\": \"custom_attribute\", \"value\": \"$opt_dummy_value\"}]",
            "id": "$opt_dummy_audience",
            "name": "Optimizely-Generated Audience for Backwards Compatibility"
        }
    ],
    "anonymizeIP": True,
    "sdkKey": "91GuiKYH8ZF1hLLXR7DR1",
    "attributes": [
        {
            "id": "23749050365",
            "key": "laptop_os_version"
        },
        {
            "id": "23767440563",
            "key": "laptop_os"
        }
    ],
    "botFiltering": False,
    "accountId": "10845721364",
    "events": [
        {
            "experimentIds": [
                "9300000235908"
            ],
            "id": "23746420417",
            "key": "myevent"
        }
    ],
    "revision": "19"
}
