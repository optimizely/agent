holdouts_datafile = {
    "accountId": "12133785640",
    "projectId": "6460519658291200",
    "revision": "12",
    "attributes": [
        {"id": "5502380200951808", "key": "all"},
        {"id": "5750214343000064", "key": "ho"}
    ],
    "audiences": [
        {
            "name": "ho_3_aud",
            "conditions": "[\"or\", {\"match\": \"exact\", \"name\": \"$opt_dummy_attribute\", \"type\": \"custom_attribute\", \"value\": \"$opt_dummy_value\"}]",
            "id": "5435551013142528"
        },
        {
            "name": "ho_6_aud",
            "conditions": "[\"or\", {\"match\": \"exact\", \"name\": \"$opt_dummy_attribute\", \"type\": \"custom_attribute\", \"value\": \"$opt_dummy_value\"}]",
            "id": "5841838209236992"
        },
        {
            "name": "ho_4_aud",
            "conditions": "[\"or\", {\"match\": \"exact\", \"name\": \"$opt_dummy_attribute\", \"type\": \"custom_attribute\", \"value\": \"$opt_dummy_value\"}]",
            "id": "6043616745881600"
        },
        {
            "name": "ho_5_aud",
            "conditions": "[\"or\", {\"match\": \"exact\", \"name\": \"$opt_dummy_attribute\", \"type\": \"custom_attribute\", \"value\": \"$opt_dummy_value\"}]",
            "id": "6410995866796032"
        },
        {
            "id": "$opt_dummy_audience",
            "name": "Optimizely-Generated Audience for Backwards Compatibility",
            "conditions": "[\"or\", {\"match\": \"exact\", \"name\": \"$opt_dummy_attribute\", \"type\": \"custom_attribute\", \"value\": \"$opt_dummy_value\"}]"
        }
    ],
    "version": "4",
    "events": [
        {"id": "6554438379241472", "experimentIds": [], "key": "event1"}
    ],
    "integrations": [],
    "holdouts": [
        {
            "id": "1673115",
            "key": "holdout_6",
            "status": "Running",
            "variations": [
                {"id": "$opt_dummy_variation_id", "key": "off", "featureEnabled": False, "variables": []}
            ],
            "trafficAllocation": [
                {"entityId": "$opt_dummy_variation_id", "endOfRange": 4000}
            ],
            "audienceIds": ["5841838209236992"],
            "audienceConditions": ["or", "5841838209236992"]
        },
        {
            "id": "1673114",
            "key": "holdout_5",
            "status": "Running",
            "variations": [
                {"id": "$opt_dummy_variation_id", "key": "off", "featureEnabled": False, "variables": []}
            ],
            "trafficAllocation": [
                {"entityId": "$opt_dummy_variation_id", "endOfRange": 2000}
            ],
            "audienceIds": ["6410995866796032"],
            "audienceConditions": ["or", "6410995866796032"]
        },
        {
            "id": "1673113",
            "key": "holdouts_4",
            "status": "Running",
            "variations": [
                {"id": "$opt_dummy_variation_id", "key": "off", "featureEnabled": False, "variables": []}
            ],
            "trafficAllocation": [
                {"entityId": "$opt_dummy_variation_id", "endOfRange": 5000}
            ],
            "audienceIds": ["6043616745881600"],
            "audienceConditions": ["or", "6043616745881600"]
        },
        {
            "id": "1673112",
            "key": "holdout_3",
            "status": "Running",
            "variations": [
                {"id": "$opt_dummy_variation_id", "key": "off", "featureEnabled": False, "variables": []}
            ],
            "trafficAllocation": [
                {"entityId": "$opt_dummy_variation_id", "endOfRange": 1000}
            ],
            "audienceIds": ["5435551013142528"],
            "audienceConditions": ["or", "5435551013142528"]
        }
    ],
    "anonymizeIP": True,
    "botFiltering": False,
    "typedAudiences": [
        {
            "name": "ho_3_aud",
            "conditions": ["and", ["or", ["or", {"match": "exact", "name": "ho", "type": "custom_attribute", "value": 3}], ["or", {"match": "le", "name": "all", "type": "custom_attribute", "value": 3}]]],
            "id": "5435551013142528"
        },
        {
            "name": "ho_6_aud",
            "conditions": ["and", ["or", ["or", {"match": "exact", "name": "ho", "type": "custom_attribute", "value": 6}], ["or", {"match": "le", "name": "all", "type": "custom_attribute", "value": 6}]]],
            "id": "5841838209236992"
        },
        {
            "name": "ho_4_aud",
            "conditions": ["and", ["or", ["or", {"match": "exact", "name": "ho", "type": "custom_attribute", "value": 4}], ["or", {"match": "le", "name": "all", "type": "custom_attribute", "value": 4}]]],
            "id": "6043616745881600"
        },
        {
            "name": "ho_5_aud",
            "conditions": ["and", ["or", ["or", {"match": "exact", "name": "ho", "type": "custom_attribute", "value": 5}], ["or", {"match": "le", "name": "all", "type": "custom_attribute", "value": 5}]]],
            "id": "6410995866796032"
        }
    ],
    "variables": [],
    "environmentKey": "production",
    "sdkKey": "BLsSFScP7tSY5SCYuKn8c",
    "featureFlags": [
        {"id": "497759", "key": "flag1", "rolloutId": "rollout-497759-631765411405174", "experimentIds": [], "variables": []},
        {"id": "497760", "key": "flag2", "rolloutId": "rollout-497760-631765411405174", "experimentIds": [], "variables": []}
    ],
    "rollouts": [
        {
            "id": "rollout-497759-631765411405174",
            "experiments": [
                {
                    "id": "default-rollout-497759-631765411405174",
                    "key": "default-rollout-497759-631765411405174",
                    "status": "Running",
                    "layerId": "rollout-497759-631765411405174",
                    "variations": [{"id": "1583341", "key": "variation_1", "featureEnabled": True, "variables": []}],
                    "trafficAllocation": [{"entityId": "1583341", "endOfRange": 10000}],
                    "forcedVariations": {},
                    "audienceIds": [],
                    "audienceConditions": []
                }
            ]
        },
        {
            "id": "rollout-497760-631765411405174",
            "experiments": [
                {
                    "id": "default-rollout-497760-631765411405174",
                    "key": "default-rollout-497760-631765411405174",
                    "status": "Running",
                    "layerId": "rollout-497760-631765411405174",
                    "variations": [{"id": "1583340", "key": "variation_2", "featureEnabled": True, "variables": []}],
                    "trafficAllocation": [{"entityId": "1583340", "endOfRange": 10000}],
                    "forcedVariations": {},
                    "audienceIds": [],
                    "audienceConditions": []
                }
            ]
        }
    ],
    "experiments": [],
    "groups": [],
    "region": "US"
}
