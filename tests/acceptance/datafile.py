datafile = {
    "accountId": "10845721364",
    "anonymizeIP": True,
    "attributes": [
        {
            "id": "16921322086",
            "key": "attr_1"
        }
    ],
    "audiences": [
        {
            "conditions": "[\"and\", [\"or\", [\"or\", {\"match\": \"exact\", \"name\": \"attr_1\", \"type\": \"custom_attribute\", \"value\": \"hola\"}]]]",
            "id": "16902921321",
            "name": "Audience1"
        },
        {
            "conditions": "[\"or\", {\"match\": \"exact\", \"name\": \"$opt_dummy_attribute\", \"type\": \"custom_attribute\", \"value\": \"$opt_dummy_value\"}]",
            "id": "$opt_dummy_audience",
            "name": "Optimizely-Generated Audience for Backwards Compatibility"
        }
    ],
    "botFiltering": False,
    "environmentKey": "production",
    "events": [
        {
            "experimentIds": [
                "16911963060",
                "9300002877087",
                "16910084756"
            ],
            "id": "16911532385",
            "key": "myevent"
        }
    ],
    "experiments": [
        {
            "audienceConditions": [
                "or",
                "16902921321"
            ],
            "audienceIds": [
                "16902921321"
            ],
            "forcedVariations": {},
            "id": "16910084756",
            "key": "feature_2_test",
            "layerId": "16933431472",
            "status": "Running",
            "trafficAllocation": [
                {
                    "endOfRange": 5000,
                    "entityId": "16925360560"
                },
                {
                    "endOfRange": 10000,
                    "entityId": "16925360560"
                }
            ],
            "variations": [
                {
                    "featureEnabled": True,
                    "id": "16925360560",
                    "key": "variation_1",
                    "variables": []
                },
                {
                    "featureEnabled": True,
                    "id": "16915611472",
                    "key": "variation_2",
                    "variables": []
                }
            ]
        },
        {
            "audienceConditions": [
                "or",
                "16902921321"
            ],
            "audienceIds": [
                "16902921321"
            ],
            "forcedVariations": {},
            "id": "16911963060",
            "key": "ab_test1",
            "layerId": "16916031507",
            "status": "Running",
            "trafficAllocation": [
                {
                    "endOfRange": 1000,
                    "entityId": "16905941566"
                },
                {
                    "endOfRange": 5000,
                    "entityId": "16905941566"
                },
                {
                    "endOfRange": 8000,
                    "entityId": "16905941566"
                },
                {
                    "endOfRange": 9000,
                    "entityId": "16905941566"
                },
                {
                    "endOfRange": 10000,
                    "entityId": "16905941566"
                }
            ],
            "variations": [
                {
                    "featureEnabled": True,
                    "id": "16905941566",
                    "key": "variation_1",
                    "variables": []
                },
                {
                    "featureEnabled": True,
                    "id": "16927770169",
                    "key": "variation_2",
                    "variables": []
                }
            ]
        },
        {
            "audienceConditions": [
                "or",
                "16902921321"
            ],
            "audienceIds": [
                "16902921321"
            ],
            "cmab": {
                "attributeIds": [
                    "16921322086"
                ],
                "trafficAllocation": 10000
            },
            "forcedVariations": {},
            "id": "9300002877087",
            "key": "cmab-rule_1",
            "layerId": "9300002131372",
            "status": "Running",
            "trafficAllocation": [],
            "variations": [
                {
                    "featureEnabled": False,
                    "id": "1579277",
                    "key": "off",
                    "variables": []
                },
                {
                    "featureEnabled": True,
                    "id": "1579278",
                    "key": "on",
                    "variables": []
                }
            ]
        }
    ],
    "featureFlags": [
        {
            "experimentIds": [
                "9300002877087"
            ],
            "id": "496419",
            "key": "cmab_flag",
            "rolloutId": "rollout-496419-16935023792",
            "variables": []
        },
        {
            "experimentIds": [],
            "id": "16907463855",
            "key": "feature_3",
            "rolloutId": "16909553406",
            "variables": []
        },
        {
            "experimentIds": [],
            "id": "16912161768",
            "key": "feature_4",
            "rolloutId": "16943340293",
            "variables": []
        },
        {
            "experimentIds": [],
            "id": "16923312421",
            "key": "feature_5",
            "rolloutId": "16917103311",
            "variables": []
        },
        {
            "experimentIds": [],
            "id": "16925981047",
            "key": "feature_1",
            "rolloutId": "16928980969",
            "variables": [
                {
                    "defaultValue": "hello",
                    "id": "16916052157",
                    "key": "str_var",
                    "type": "string"
                },
                {
                    "defaultValue": "5.6",
                    "id": "16923002469",
                    "key": "double_var",
                    "type": "double"
                },
                {
                    "defaultValue": "true",
                    "id": "16932993089",
                    "key": "bool_var",
                    "type": "boolean"
                },
                {
                    "defaultValue": "1",
                    "id": "16937161477",
                    "key": "int_var",
                    "type": "integer"
                }
            ]
        },
        {
            "experimentIds": [
                "16910084756"
            ],
            "id": "16928980973",
            "key": "feature_2",
            "rolloutId": "16917900798",
            "variables": []
        },
        {
            "experimentIds": [
                "16911963060"
            ],
            "id": "147680",
            "key": "GkbzTurBWXr8EtNGZj2j6e",
            "rolloutId": "rollout-147680-16935023792",
            "variables": []
        }
    ],
    "groups": [],
    "integrations": [],
    "projectId": "16931203314",
    "revision": "139",
    "rollouts": [
        {
            "experiments": [
                {
                    "audienceConditions": [],
                    "audienceIds": [],
                    "forcedVariations": {},
                    "id": "default-rollout-496419-16935023792",
                    "key": "default-rollout-496419-16935023792",
                    "layerId": "rollout-496419-16935023792",
                    "status": "Running",
                    "trafficAllocation": [
                        {
                            "endOfRange": 10000,
                            "entityId": "1579279"
                        }
                    ],
                    "variations": [
                        {
                            "featureEnabled": False,
                            "id": "1579279",
                            "key": "off",
                            "variables": []
                        }
                    ]
                }
            ],
            "id": "rollout-496419-16935023792"
        },
        {
            "experiments": [
                {
                    "audienceConditions": [],
                    "audienceIds": [],
                    "forcedVariations": {},
                    "id": "default-16909553406",
                    "key": "default-16909553406",
                    "layerId": "16909553406",
                    "status": "Running",
                    "trafficAllocation": [
                        {
                            "endOfRange": 10000,
                            "entityId": "471185"
                        }
                    ],
                    "variations": [
                        {
                            "featureEnabled": False,
                            "id": "471185",
                            "key": "off",
                            "variables": []
                        }
                    ]
                }
            ],
            "id": "16909553406"
        },
        {
            "experiments": [
                {
                    "audienceConditions": [],
                    "audienceIds": [],
                    "forcedVariations": {},
                    "id": "default-16943340293",
                    "key": "default-16943340293",
                    "layerId": "16943340293",
                    "status": "Running",
                    "trafficAllocation": [
                        {
                            "endOfRange": 10000,
                            "entityId": "16925940659"
                        }
                    ],
                    "variations": [
                        {
                            "featureEnabled": True,
                            "id": "16925940659",
                            "key": "16925940659",
                            "variables": []
                        }
                    ]
                }
            ],
            "id": "16943340293"
        },
        {
            "experiments": [
                {
                    "audienceConditions": [],
                    "audienceIds": [],
                    "forcedVariations": {},
                    "id": "default-16917103311",
                    "key": "default-16917103311",
                    "layerId": "16917103311",
                    "status": "Running",
                    "trafficAllocation": [
                        {
                            "endOfRange": 10000,
                            "entityId": "16927890136"
                        }
                    ],
                    "variations": [
                        {
                            "featureEnabled": True,
                            "id": "16927890136",
                            "key": "16927890136",
                            "variables": []
                        }
                    ]
                }
            ],
            "id": "16917103311"
        },
        {
            "experiments": [
                {
                    "audienceConditions": [
                        "or",
                        "16902921321"
                    ],
                    "audienceIds": [
                        "16902921321"
                    ],
                    "forcedVariations": {},
                    "id": "16941022436",
                    "key": "16941022436",
                    "layerId": "16928980969",
                    "status": "Running",
                    "trafficAllocation": [
                        {
                            "endOfRange": 10000,
                            "entityId": "16906801184"
                        }
                    ],
                    "variations": [
                        {
                            "featureEnabled": True,
                            "id": "16906801184",
                            "key": "16906801184",
                            "variables": [
                                {
                                    "id": "16916052157",
                                    "value": "hello"
                                },
                                {
                                    "id": "16923002469",
                                    "value": "5.6"
                                },
                                {
                                    "id": "16932993089",
                                    "value": "true"
                                },
                                {
                                    "id": "16937161477",
                                    "value": "1"
                                }
                            ]
                        }
                    ]
                },
                {
                    "audienceConditions": [],
                    "audienceIds": [],
                    "forcedVariations": {},
                    "id": "default-16928980969",
                    "key": "default-16928980969",
                    "layerId": "16928980969",
                    "status": "Running",
                    "trafficAllocation": [
                        {
                            "endOfRange": 10000,
                            "entityId": "471188"
                        }
                    ],
                    "variations": [
                        {
                            "featureEnabled": False,
                            "id": "471188",
                            "key": "off",
                            "variables": [
                                {
                                    "id": "16916052157",
                                    "value": "hello"
                                },
                                {
                                    "id": "16923002469",
                                    "value": "5.6"
                                },
                                {
                                    "id": "16932993089",
                                    "value": "true"
                                },
                                {
                                    "id": "16937161477",
                                    "value": "1"
                                }
                            ]
                        }
                    ]
                }
            ],
            "id": "16928980969"
        },
        {
            "experiments": [
                {
                    "audienceConditions": [
                        "or",
                        "16902921321"
                    ],
                    "audienceIds": [
                        "16902921321"
                    ],
                    "forcedVariations": {},
                    "id": "16924931120",
                    "key": "16924931120",
                    "layerId": "16917900798",
                    "status": "Running",
                    "trafficAllocation": [
                        {
                            "endOfRange": 10000,
                            "entityId": "16931381940"
                        }
                    ],
                    "variations": [
                        {
                            "featureEnabled": True,
                            "id": "16931381940",
                            "key": "16931381940",
                            "variables": []
                        }
                    ]
                },
                {
                    "audienceConditions": [],
                    "audienceIds": [],
                    "forcedVariations": {},
                    "id": "default-16917900798",
                    "key": "default-16917900798",
                    "layerId": "16917900798",
                    "status": "Running",
                    "trafficAllocation": [
                        {
                            "endOfRange": 10000,
                            "entityId": "471189"
                        }
                    ],
                    "variations": [
                        {
                            "featureEnabled": False,
                            "id": "471189",
                            "key": "off",
                            "variables": []
                        }
                    ]
                }
            ],
            "id": "16917900798"
        },
        {
            "experiments": [
                {
                    "audienceConditions": [],
                    "audienceIds": [],
                    "forcedVariations": {},
                    "id": "default-rollout-147680-16935023792",
                    "key": "default-rollout-147680-16935023792",
                    "layerId": "rollout-147680-16935023792",
                    "status": "Running",
                    "trafficAllocation": [
                        {
                            "endOfRange": 10000,
                            "entityId": "471190"
                        }
                    ],
                    "variations": [
                        {
                            "featureEnabled": False,
                            "id": "471190",
                            "key": "off",
                            "variables": []
                        }
                    ]
                }
            ],
            "id": "rollout-147680-16935023792"
        }
    ],
    "sdkKey": "KZbunNn9bVfBWLpZPq2XC4",
    "typedAudiences": [],
    "variables": [],
    "version": "4"
}
