import json
import os

import pytest
import requests

from tests.acceptance.helpers import ENDPOINT_DECIDE
from tests.acceptance.helpers import create_and_validate_request_and_response
from tests.acceptance.helpers import sort_response

BASE_URL = os.getenv('host')

expected_single_flag_key = """
    {
      "variationKey": "variation_1",
      "enabled": true,
      "ruleKey": "feature_2_test",
      "flagKey": "feature_2",
      "userContext": {
        "userId": "matjaz",
        "attributes": {
          "attr_1": "hola"
        }
      },
      "reasons": ["Audiences for experiment feature_2_test collectively evaluated to true."]
    }
"""

expected_invalid_flag_key = r"""
    {
      "variationKey": "",
      "enabled": false,
      "ruleKey": "",
      "flagKey": "invalid_flag_key",
      "userContext": {
        "userId": "matjaz",
        "attributes": {
          "attr_1": "hola"
        }
      },
      "reasons": [
        "No flag was found for key \"[invalid_flag_key]\"."
      ]
    }
"""


@pytest.mark.parametrize(
    "flag_key, expected_response, expected_status_code", [
        ("feature_2", expected_single_flag_key, 200),
        ("invalid_flag_key", expected_invalid_flag_key, 200),
    ],
    ids=["valid case", "invalid_flag_key"])
def test_decide__feature(session_obj, flag_key, expected_response, expected_status_code):
    """
    Test validates:
    Correct response when valid and invalid flag key are passed as parameters.
    ...
    :param session_obj:
    :param flag_key:
    :param expected_response:
    :param expected_status_code:
    """
    payload = """
        {
          "userId": "matjaz",
          "decideOptions": [
              "ENABLED_FLAGS_ONLY",
              "INCLUDE_REASONS"
          ],
          "userAttributes": {"attr_1": "hola"}
        }
    """

    params = {"keys": flag_key}
    resp = create_and_validate_request_and_response(ENDPOINT_DECIDE, 'post', session_obj, payload=payload,
                                                    params=params)

    assert json.loads(expected_response) == resp.json()
    assert resp.status_code == expected_status_code, resp.text
    resp.raise_for_status()


expected_flag_keys = r"""
[
  {
    "variationKey": "16925940659",
    "enabled": true,
    "ruleKey": "16939051724",
    "flagKey": "feature_4",
    "userContext": {
      "userId": "matjaz",
      "attributes": {
        "attr_1": "hola"
      }
    },
    "reasons": [
      "Audiences for experiment 16939051724 collectively evaluated to true.",
      "User \"matjaz\" meets conditions for targeting rule \"Everyone Else\"."
    ]
  },
  {
    "variationKey": "16927890136",
    "enabled": true,
    "ruleKey": "16932940705",
    "flagKey": "feature_5",
    "userContext": {
      "userId": "matjaz",
      "attributes": {
        "attr_1": "hola"
      }
    },
    "reasons": [
      "Audiences for experiment 16932940705 collectively evaluated to true.",
      "User \"matjaz\" meets conditions for targeting rule \"Everyone Else\"."
    ]
  },
  {
    "variationKey": "16906801184",
    "enabled": true,
    "ruleKey": "16941022436",
    "flagKey": "feature_1",
    "userContext": {
      "userId": "matjaz",
      "attributes": {
        "attr_1": "hola"
      }
    },
    "reasons": [
      "Audiences for experiment 16941022436 collectively evaluated to true.",
      "User \"matjaz\" meets conditions for targeting rule \"Everyone Else\"."
    ],
    "variables": {
      "bool_var": true,
      "double_var": 5.6,
      "int_var": 1,
      "str_var": "hello"
    }
  },
  {
    "variationKey": "variation_1",
    "enabled": true,
    "ruleKey": "feature_2_test",
    "flagKey": "feature_2",
    "userContext": {
      "userId": "matjaz",
      "attributes": {
        "attr_1": "hola"
      }
    },
    "reasons": [
      "Audiences for experiment feature_2_test collectively evaluated to true."
    ]
  }
]
"""


@pytest.mark.parametrize(
    "parameters, expected_response, expected_status_code", [
        ({}, expected_flag_keys, 200),
        ({"keys": []}, expected_flag_keys, 200),
        ({"keys": ["feature_1", "feature_2", "feature_4", "feature_5"]}, expected_flag_keys, 200),
    ],
    ids=["missig_flagkey_parameter", "no flag key specified", "multiple flag keys"])
def test_decide__flag_key_parameter(session_obj, parameters, expected_response, expected_status_code):
    """
    Test validates:
    That no required parameter, empty param and all parameters return identical response.
    Openapi spec specifies 400 for missing flagKey parameter. But We keep 400 status code in the openapi spec
    for missing reuired parameter, even though when no flagKey parameter is supplied to the request,
    Agent still responds with all decisions and status 200.
    That is consistent with the behavior of activate and other api-s
    :param session_obj:
    :param parameters:
    :param expected_response:
    :param expected_status_code:
    """
    payload = """
        {
          "userId": "matjaz",
          "decideOptions": [
              "ENABLED_FLAGS_ONLY",
              "INCLUDE_REASONS"
          ],
          "userAttributes": {"attr_1": "hola"}
        }
    """

    params = parameters
    resp = create_and_validate_request_and_response(ENDPOINT_DECIDE, 'post', session_obj, payload=payload,
                                                    params=params)

    sorted_actual = sort_response(resp.json(), 'flagKey')
    sorted_expected = sort_response(json.loads(expected_response), 'flagKey')
    assert sorted_actual == sorted_expected


def test_decide_403(session_override_sdk_key):
    """
    Test that 403 Forbidden is returned. We use invalid SDK key to trigger 403.
    :param session_override_sdk_key: sdk key to override the session using invalid sdk key
    """
    payload = '{"userId": "matjaz", "userAttributes": {"attr_1": "hola"}}'
    params = {"flagKey": "feature_2"}

    with pytest.raises(requests.exceptions.HTTPError):
        resp = create_and_validate_request_and_response(ENDPOINT_DECIDE, 'post', session_override_sdk_key,
                                                        payload=payload, params=params)

        assert resp.status_code == 403
        assert resp.json()['error'] == 'unable to fetch fresh datafile (consider ' \
                                       'rechecking SDK key), status code: 403 Forbidden'

        resp.raise_for_status()
