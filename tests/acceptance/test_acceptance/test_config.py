import json
import os
import pytest
import requests

from tests.acceptance.helpers import ENDPOINT_CONFIG
from tests.acceptance.helpers import create_and_validate_request
from tests.acceptance.helpers import create_and_validate_response
from tests.acceptance.helpers import create_and_validate_request_and_response

BASE_URL = os.getenv('host')

expected_config = """{
    "revision": "111",
    "experimentsMap": {
        "ab_test1": {
            "id": "16911963060",
            "key": "ab_test1",
            "variationsMap": {
                "variation_1": {
                    "id": "16905941566",
                    "key": "variation_1",
                    "featureEnabled": false,
                    "variablesMap": {}
                },
                "variation_2": {
                    "id": "16927770169",
                    "key": "variation_2",
                    "featureEnabled": false,
                    "variablesMap": {}
                }
            }
        },
        "feature_2_test": {
            "id": "16910084756",
            "key": "feature_2_test",
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
        "feature_1": {
            "id": "16925981047",
            "key": "feature_1",
            "experimentsMap": {},
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
        },
        "feature_2": {
            "id": "16928980973",
            "key": "feature_2",
            "experimentsMap": {
                "feature_2_test": {
                    "id": "16910084756",
                    "key": "feature_2_test",
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
            "variablesMap": {}
        },
        "feature_3": {
            "id": "16907463855",
            "key": "feature_3",
            "experimentsMap": {},
            "variablesMap": {}
        },
        "feature_4": {
            "id": "16912161768",
            "key": "feature_4",
            "experimentsMap": {},
            "variablesMap": {}
        },
        "feature_5": {
            "id": "16923312421",
            "key": "feature_5",
            "experimentsMap": {},
            "variablesMap": {}
        }
    }
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
