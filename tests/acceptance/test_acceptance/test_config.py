import json

import pytest
import requests

from tests.acceptance.helpers import ENDPOINT_CONFIG
from tests.acceptance.helpers import create_and_validate_request_and_response

expected_config = """{
    "environmentKey": "production",
    "sdkKey": "KZbunNn9bVfBWLpZPq2XC4",
    "revision": "139",
    "experimentsMap": {
        "ab_test1": {
            "id": "16911963060",
            "key": "ab_test1",
            "audiences": "\"Audience1\"",
            "variationsMap": {
                "variation_1": {
                    "id": "16905941566",
                    "key": "variation_1",
                    "featureEnabled": true,
                    "variablesMap": {}
                },
                "variation_2": {
                    "id": "16927770169",
                    "key": "variation_2",
                    "featureEnabled": true,
                    "variablesMap": {}
                }
            }
        },
        "cmab-rule_1": {
            "id": "9300002877087",
            "key": "cmab-rule_1",
            "audiences": "Audience1",
            "variationsMap": {
                "off": {
                    "id": "1579277",
                    "key": "off",
                    "featureEnabled": false,
                    "variablesMap": {}
                },
                "on": {
                    "id": "1579278",
                    "key": "on",
                    "featureEnabled": true,
                    "variablesMap": {}
                }
            }
        },
        "feature_2_test": {
            "id": "16910084756",
            "key": "feature_2_test",
            "audiences": "Audience1",
            "variationsMap": {
                "variation_1": {
                    "id": "16925360560",
                    "key": "variation_1",
                    "featureEnabled": true,
                    "variablesMap": {}
                },
                "variation_2": {
                    "id": "16915611472",
                    "key": "variation_2",
                    "featureEnabled": true,
                    "variablesMap": {}
                }
            }
        }
    },
    "featuresMap": {
        "GkbzTurBWXr8EtNGZj2j6e": {
            "id": "147680",
            "key": "GkbzTurBWXr8EtNGZj2j6e",
            "experimentRules": [
                {
                    "id": "16911963060",
                    "key": "ab_test1",
                    "audiences": "Audience1",
                    "variationsMap": {
                        "variation_1": {
                            "id": "16905941566",
                            "key": "variation_1",
                            "featureEnabled": true,
                            "variablesMap": {}
                        },
                        "variation_2": {
                            "id": "16927770169",
                            "key": "variation_2",
                            "featureEnabled": true,
                            "variablesMap": {}
                        }
                    }
                }
            ],
            "deliveryRules": [
                {
                    "id": "default-rollout-147680-16935023792",
                    "key": "default-rollout-147680-16935023792",
                    "audiences": "",
                    "variationsMap": {
                        "off": {
                            "id": "471190",
                            "key": "off",
                            "featureEnabled": false,
                            "variablesMap": {}
                        }
                    }
                }
            ],
            "variablesMap": {},
            "experimentsMap": {
                "ab_test1": {
                    "id": "16911963060",
                    "key": "ab_test1",
                    "audiences": "Audience1",
                    "variationsMap": {
                        "variation_1": {
                            "id": "16905941566",
                            "key": "variation_1",
                            "featureEnabled": true,
                            "variablesMap": {}
                        },
                        "variation_2": {
                            "id": "16927770169",
                            "key": "variation_2",
                            "featureEnabled": true,
                            "variablesMap": {}
                        }
                    }
                }
            }
        },
        "feature_1": {
            "id": "16925981047",
            "key": "feature_1",
            "experimentRules": [],
            "deliveryRules": [
                {
                    "id": "16941022436",
                    "key": "16941022436",
                    "audiences": "Audience1",
                    "variationsMap": {
                        "16906801184": {
                            "id": "16906801184",
                            "key": "16906801184",
                            "featureEnabled": true,
                            "variablesMap": {
                                "bool_var": {
                                    "id": "16932993089",
                                    "key": "bool_var",
                                    "type": "boolean",
                                    "value": "true"
                                },
                                "double_var": {
                                    "id": "16923002469",
                                    "key": "double_var",
                                    "type": "double",
                                    "value": "5.6"
                                },
                                "int_var": {
                                    "id": "16937161477",
                                    "key": "int_var",
                                    "type": "integer",
                                    "value": "1"
                                },
                                "str_var": {
                                    "id": "16916052157",
                                    "key": "str_var",
                                    "type": "string",
                                    "value": "hello"
                                }
                            }
                        }
                    }
                },
                {
                    "id": "default-16928980969",
                    "key": "default-16928980969",
                    "audiences": "",
                    "variationsMap": {
                        "off": {
                            "id": "471188",
                            "key": "off",
                            "featureEnabled": false,
                            "variablesMap": {
                                "bool_var": {
                                    "id": "16932993089",
                                    "key": "bool_var",
                                    "type": "boolean",
                                    "value": "true"
                                },
                                "double_var": {
                                    "id": "16923002469",
                                    "key": "double_var",
                                    "type": "double",
                                    "value": "5.6"
                                },
                                "int_var": {
                                    "id": "16937161477",
                                    "key": "int_var",
                                    "type": "integer",
                                    "value": "1"
                                },
                                "str_var": {
                                    "id": "16916052157",
                                    "key": "str_var",
                                    "type": "string",
                                    "value": "hello"
                                }
                            }
                        }
                    }
                }
            ],
            "variablesMap": {
                "bool_var": {
                    "id": "16932993089",
                    "key": "bool_var",
                    "type": "boolean",
                    "value": "true"
                },
                "double_var": {
                    "id": "16923002469",
                    "key": "double_var",
                    "type": "double",
                    "value": "5.6"
                },
                "int_var": {
                    "id": "16937161477",
                    "key": "int_var",
                    "type": "integer",
                    "value": "1"
                },
                "str_var": {
                    "id": "16916052157",
                    "key": "str_var",
                    "type": "string",
                    "value": "hello"
                }
            },
            "experimentsMap": {}
        },
        "feature_2": {
            "id": "16928980973",
            "key": "feature_2",
            "experimentRules": [
                {
                    "id": "16910084756",
                    "key": "feature_2_test",
                    "audiences": "Audience1",
                    "variationsMap": {
                        "variation_1": {
                            "id": "16925360560",
                            "key": "variation_1",
                            "featureEnabled": true,
                            "variablesMap": {}
                        },
                        "variation_2": {
                            "id": "16915611472",
                            "key": "variation_2",
                            "featureEnabled": true,
                            "variablesMap": {}
                        }
                    }
                }
            ],
            "deliveryRules": [
                {
                    "id": "16924931120",
                    "key": "16924931120",
                    "audiences": "Audience1",
                    "variationsMap": {
                        "16931381940": {
                            "id": "16931381940",
                            "key": "16931381940",
                            "featureEnabled": true,
                            "variablesMap": {}
                        }
                    }
                },
                {
                    "id": "default-16917900798",
                    "key": "default-16917900798",
                    "audiences": "",
                    "variationsMap": {
                        "off": {
                            "id": "471189",
                            "key": "off",
                            "featureEnabled": false,
                            "variablesMap": {}
                        }
                    }
                }
            ],
            "variablesMap": {},
            "experimentsMap": {
                "feature_2_test": {
                    "id": "16910084756",
                    "key": "feature_2_test",
                    "audiences": "Audience1",
                    "variationsMap": {
                        "variation_1": {
                            "id": "16925360560",
                            "key": "variation_1",
                            "featureEnabled": true,
                            "variablesMap": {}
                        },
                        "variation_2": {
                            "id": "16915611472",
                            "key": "variation_2",
                            "featureEnabled": true,
                            "variablesMap": {}
                        }
                    }
                }
            }
        },
        "feature_3": {
            "id": "16907463855",
            "key": "feature_3",
            "experimentRules": [],
            "deliveryRules": [
                {
                    "id": "default-16909553406",
                    "key": "default-16909553406",
                    "audiences": "",
                    "variationsMap": {
                        "off": {
                            "id": "471185",
                            "key": "off",
                            "featureEnabled": false,
                            "variablesMap": {}
                        }
                    }
                }
            ],
            "variablesMap": {},
            "experimentsMap": {}
        },
        "feature_4": {
            "id": "16912161768",
            "key": "feature_4",
            "experimentRules": [],
            "deliveryRules": [
                {
                    "id": "default-16943340293",
                    "key": "default-16943340293",
                    "audiences": "",
                    "variationsMap": {
                        "16925940659": {
                            "id": "16925940659",
                            "key": "16925940659",
                            "featureEnabled": true,
                            "variablesMap": {}
                        }
                    }
                }
            ],
            "variablesMap": {},
            "experimentsMap": {}
        },
        "feature_5": {
            "id": "16923312421",
            "key": "feature_5",
            "experimentRules": [],
            "deliveryRules": [
                {
                    "id": "default-16917103311",
                    "key": "default-16917103311",
                    "audiences": "",
                    "variationsMap": {
                        "16927890136": {
                            "id": "16927890136",
                            "key": "16927890136",
                            "featureEnabled": true,
                            "variablesMap": {}
                        }
                    }
                }
            ],
            "variablesMap": {},
            "experimentsMap": {}
        }
    },
    "attributes": [
        {
            "id": "16921322086",
            "key": "attr_1"
        }
    ],
    "audiences": [
        {
            "id": "16902921321",
            "name": "Audience1",
            "conditions": "[\"and\", [\"or\", [\"or\", {\"match\": \"exact\", \"name\": \"attr_1\", \"type\": \"custom_attribute\", \"value\": \"hola\"}]]]"
        }
    ],
    "events": [
        {
            "id": "16911532385",
            "key": "myevent",
            "experimentIds": [
                "16911963060",
                "9300002877087",
                "16910084756"
            ]
        }
    ]
}"""


def test_config(session_obj):
    """
    Test validates all returned available experiment and features definitions
    for this environment.

    Note: Test will fail as soon as anything in the response body is modified.
    If someone updates any of the fields, the expected_response will need to be updated
    as well.
    :param session_obj: session object
    """
    resp = create_and_validate_request_and_response(ENDPOINT_CONFIG, 'get', session_obj)

    assert resp.status_code == 200
    resp.raise_for_status()
    assert json.loads(expected_config) == resp.json()


def test_config_403(session_override_sdk_key):
    """
    Test that 403 Forbidden is returned. We use invalid SDK key to trigger 403.
    :param : session_obj
    """
    with pytest.raises(requests.exceptions.HTTPError):
        resp = create_and_validate_request_and_response(ENDPOINT_CONFIG, 'get', session_override_sdk_key)

        assert resp.status_code == 403
        assert resp.json()['error'] == 'unable to fetch fresh datafile (consider ' \
                                       'rechecking SDK key), status code: 403 Forbidden'

        resp.raise_for_status()
