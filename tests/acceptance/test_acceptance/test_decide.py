import json

import pytest
import requests

from tests.acceptance.helpers import ENDPOINT_DECIDE
from tests.acceptance.helpers import create_and_validate_request_and_response
from tests.acceptance.helpers import sort_response

expected_forced_decision_without_rule_key = {
    "variationKey": "variation_1",
    "enabled": True,
    "ruleKey": "",
    "flagKey": "feature_2",
    "userContext": {
        "userId": "matjaz",
        "attributes": {
            "attr_1": "hola"
        }
    },
    "reasons": ["Variation (variation_1) is mapped to flag (feature_2) and user (matjaz) in the forced decision map."]
}

expected_forced_decision_with_rule_key = {
    "variationKey": "variation_2",
    "enabled": True,
    "ruleKey": "feature_2_test",
    "flagKey": "feature_2",
    "userContext": {
        "userId": "matjaz",
        "attributes": {
            "attr_1": "hola"
        }
    },
    "reasons": [
        "Variation (variation_2) is mapped to flag (feature_2), rule (feature_2_test) and user (matjaz) "
        "in the forced decision map."]
}

expected_single_flag_key_with_ups = r"""
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
      "reasons": ["User \"matjaz\" was previously bucketed into variation \"variation_1\" of experiment \"feature_2_test\"."]
    }
"""

expected_single_flag_key_no_ups = r"""
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
        "No flag was found for key \"invalid_flag_key\"."
      ]
    }
"""


@pytest.mark.parametrize(
    "flag_key, expected_response, expected_status_code", [
        ("feature_2", expected_single_flag_key_with_ups, 200),
        ("invalid_flag_key", expected_invalid_flag_key, 200),
    ],
    ids=["valid case with ups", "invalid_flag_key"])
def test_decide__feature_with_ups(session_obj, flag_key, expected_response, expected_status_code):
    """
    Test validates:
    Correct response for flag key when User profile service is enabled.
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
    resp = create_and_validate_request_and_response(ENDPOINT_DECIDE, "post", session_obj, payload=payload,
                                                    params=params)

    # assert that expected response includes sticky variation from UPS (expected_single_flag_key_ups)
    assert json.loads(expected_response) == resp.json()
    assert resp.status_code == expected_status_code, resp.text
    resp.raise_for_status()


@pytest.mark.parametrize(
    "flag_key, expected_response, expected_status_code", [
        ("feature_2", expected_single_flag_key_no_ups, 200),
        ("invalid_flag_key", expected_invalid_flag_key, 200),
    ],
    ids=["valid case no ups", "invalid_flag_key"])
def test_decide__feature_no_ups(session_obj, flag_key, expected_response, expected_status_code):
    """
    Test validates:
    Correct response for flag key when User profile service is not enabled.
    This test is required to be run on Agent on Amazon Web Services.
    It is only used there. And it is excluded from the test run in this repo.
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
    resp = create_and_validate_request_and_response(ENDPOINT_DECIDE, "post", session_obj, payload=payload,
                                                    params=params)

    # assert that expected response doesn't include sticky variation from UPS (should be expected_single_flag_key)
    assert json.loads(expected_response) == resp.json()
    assert resp.status_code == expected_status_code, resp.text
    resp.raise_for_status()


expected_flag_keys_with_ups = r"""[
  {
    "variationKey": "16925940659",
    "enabled": true,
    "ruleKey": "16939051724",
    "flagKey": "feature_4",
      "userContext": {"userId": "matjaz", "attributes": {"attr_1": "hola"}},
      "reasons": ["Audiences for experiment 16939051724 collectively evaluated to true."]},
  {
    "variationKey": "variation_1",
    "enabled": true,
    "ruleKey": "ab_test1",
    "flagKey": "flag_ab_test1",
      "userContext": {"userId": "matjaz", "attributes": {"attr_1": "hola"}},
      "reasons": ["User \"matjaz\" was previously bucketed into variation \"variation_1\" of experiment \"ab_test1\"."]},
  {
    "variationKey": "variation_1",
    "enabled": true,
    "ruleKey": "feature_2_test",
    "flagKey": "feature_2",
      "userContext": {"userId": "matjaz", "attributes": {"attr_1": "hola"}},
    "reasons": ["User \"matjaz\" was previously bucketed into variation \"variation_1\" of experiment \"feature_2_test\"."]},
  {
    "variationKey": "16927890136",
    "enabled": true,
    "ruleKey": "16932940705",
    "flagKey": "feature_5",
      "userContext": {"userId": "matjaz", "attributes": {"attr_1": "hola"}},
    "reasons": ["Audiences for experiment 16932940705 collectively evaluated to true."]},
  {
    "variationKey": "16906801184",
    "enabled": true,
    "ruleKey": "16941022436",
    "flagKey": "feature_1",
      "userContext": {"userId": "matjaz", "attributes": {"attr_1": "hola"}},
    "reasons": ["Audiences for experiment 16941022436 collectively evaluated to true."],
    "variables": {"bool_var": true, "double_var": 5.6, "int_var": 1, "str_var": "hello"}
  }
]"""

expected_flag_key__multiple_parameters_with_ups = r"""[
    {
      "variationKey": "16906801184",
      "enabled": true,
      "ruleKey": "16941022436",
      "flagKey": "feature_1",
        "userContext": {"userId": "matjaz", "attributes": {"attr_1": "hola"}},
      "reasons": ["Audiences for experiment 16941022436 collectively evaluated to true."],
      "variables": {"bool_var": true, "double_var": 5.6, "int_var": 1, "str_var": "hello"}},
    {
      "variationKey": "variation_1",
      "enabled": true,
      "ruleKey": "feature_2_test",
      "flagKey": "feature_2",
        "userContext": {"userId": "matjaz", "attributes": {"attr_1": "hola"}},
      "reasons": ["User \"matjaz\" was previously bucketed into variation \"variation_1\" of experiment \"feature_2_test\"."]},
    {
      "variationKey": "16925940659",
      "enabled": true,
      "ruleKey": "16939051724",
      "flagKey": "feature_4",
        "userContext": {"userId": "matjaz", "attributes": {"attr_1": "hola"}},
      "reasons": ["Audiences for experiment 16939051724 collectively evaluated to true."]},
    {
      "variationKey": "16927890136",
      "enabled": true,
      "ruleKey": "16932940705",
      "flagKey": "feature_5",
        "userContext": {"userId": "matjaz", "attributes": {"attr_1": "hola"}},
      "reasons": ["Audiences for experiment 16932940705 collectively evaluated to true."]
    }
]"""

expected_flag_keys_no_ups = r"""[
  {
    "variationKey": "16925940659",
    "enabled": true,
    "ruleKey": "16939051724",
    "flagKey": "feature_4",
      "userContext": {"userId": "matjaz", "attributes": {"attr_1": "hola"}},
      "reasons": ["Audiences for experiment 16939051724 collectively evaluated to true."]},
  {
    "variationKey": "variation_1",
    "enabled": true,
    "ruleKey": "ab_test1",
    "flagKey": "GkbzTurBWXr8EtNGZj2j6e",
      "userContext": {"userId": "matjaz", "attributes": {"attr_1": "hola"}},
      "reasons": ["Audiences for experiment ab_test1 collectively evaluated to true."]},
  {
    "variationKey": "variation_1",
    "enabled": true,
    "ruleKey": "feature_2_test",
    "flagKey": "feature_2",
      "userContext": {"userId": "matjaz", "attributes": {"attr_1": "hola"}},
    "reasons": ["Audiences for experiment feature_2_test collectively evaluated to true."]},
  {
    "variationKey": "16927890136",
    "enabled": true,
    "ruleKey": "16932940705",
    "flagKey": "feature_5",
      "userContext": {"userId": "matjaz", "attributes": {"attr_1": "hola"}},
    "reasons": ["Audiences for experiment 16932940705 collectively evaluated to true."]},
  {
    "variationKey": "16906801184",
    "enabled": true,
    "ruleKey": "16941022436",
    "flagKey": "feature_1",
      "userContext": {"userId": "matjaz", "attributes": {"attr_1": "hola"}},
    "reasons": ["Audiences for experiment 16941022436 collectively evaluated to true."],
    "variables": {"bool_var": true, "double_var": 5.6, "int_var": 1, "str_var": "hello"}
  }
]"""

expected_flag_key__multiple_parameters_no_ups = r"""[
    {
      "variationKey": "16906801184",
      "enabled": true,
      "ruleKey": "16941022436",
      "flagKey": "feature_1",
        "userContext": {"userId": "matjaz", "attributes": {"attr_1": "hola"}},
      "reasons": ["Audiences for experiment 16941022436 collectively evaluated to true."],
      "variables": {"bool_var": true, "double_var": 5.6, "int_var": 1, "str_var": "hello"}},
    {
      "variationKey": "variation_1",
      "enabled": true,
      "ruleKey": "feature_2_test",
      "flagKey": "feature_2",
        "userContext": {"userId": "matjaz", "attributes": {"attr_1": "hola"}},
      "reasons": ["Audiences for experiment feature_2_test collectively evaluated to true."]},
    {
      "variationKey": "16925940659",
      "enabled": true,
      "ruleKey": "default-16943340293",
      "flagKey": "feature_4",
        "userContext": {"userId": "matjaz", "attributes": {"attr_1": "hola"}},
      "reasons": ["Audiences for experiment default-16943340293 collectively evaluated to true."]},
    {
      "variationKey": "16927890136",
      "enabled": true,
      "ruleKey": "default-16917103311",
      "flagKey": "feature_5",
        "userContext": {"userId": "matjaz", "attributes": {"attr_1": "hola"}},
      "reasons": ["Audiences for experiment default-16917103311 collectively evaluated to true."]
    }
]"""


@pytest.mark.parametrize(
    "parameters, expected_response, expected_status_code, bypass_validation_request, bypass_validation_response", [
        ({}, expected_flag_keys_with_ups, 200, True, True),
        ({"keys": []}, expected_flag_keys_with_ups, 200, True, True),
        ({"keys": ["feature_1", "feature_2", "feature_4", "feature_5"]},
         expected_flag_key__multiple_parameters_with_ups, 200, True, True),
    ],
    ids=["missig_flagkey_parameter", "no flag key specified", "multiple parameters"])
def test_decide__flag_key_parameter_with_ups(session_obj, parameters, expected_response, expected_status_code,
                                             bypass_validation_request,
                                             bypass_validation_response):
    """
    Test validates:
    That no required parameter and empty param return identical response.
    Openapi spec specifies 400 for missing flagKey parameter. But We keep 400 status code in the openapi spec
    for missing reuired parameter, even though when no flagKey parameter is supplied to the request,
    Agent still responds with all decisions and status 200.
    That is consistent with the behavior of activate and other api-s
    :param session_obj: session obj
    :param parameters:  sesison obj, params, expected, expected status code
    :param expected_response: expected_flag_keys
    :param expected_status_code: 200
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
    resp = create_and_validate_request_and_response(ENDPOINT_DECIDE, "post", session_obj, bypass_validation_request,
                                                    bypass_validation_response,
                                                    payload=payload,
                                                    params=params)

    sorted_actual = sort_response(resp.json(), "flagKey")
    sorted_expected = sort_response(json.loads(expected_response), "flagKey")

    assert sorted_actual == sorted_expected


@pytest.mark.parametrize(
    "parameters, expected_response, expected_status_code, bypass_validation_request, bypass_validation_response", [
        ({}, expected_flag_keys_no_ups, 200, True, True),
        ({"keys": []}, expected_flag_keys_no_ups, 200, True, True),
        ({"keys": ["feature_1", "feature_2", "feature_4", "feature_5"]},
         expected_flag_key__multiple_parameters_no_ups, 200, True, True),
    ],
    ids=["missig_flagkey_parameter_no_ups", "no flag key specified_no_ups", "multiple parameters_no_ups"])
def test_decide__flag_key_parameter_no_ups(session_obj, parameters, expected_response, expected_status_code,
                                           bypass_validation_request,
                                           bypass_validation_response):
    """
    This test is required to be run on Agent on Amazon Web Services.
    It is only used there. And it is excluded from the test run in this repo.
    :param session_obj: session obj
    :param parameters:  sesison obj, params, expected, expected status code
    :param expected_response: expected_flag_keys
    :param expected_status_code: 200
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
    resp = create_and_validate_request_and_response(ENDPOINT_DECIDE, "post", session_obj, bypass_validation_request,
                                                    bypass_validation_response,
                                                    payload=payload,
                                                    params=params)

    sorted_actual = sort_response(resp.json(), "flagKey")
    sorted_expected = sort_response(json.loads(expected_response), "flagKey")

    assert sorted_actual == sorted_expected


@pytest.mark.parametrize(
    "flag_key, expected_response, expected_status_code, forced_flag, forced_rule, forced_variation", [
        ("feature_2", expected_forced_decision_without_rule_key, 200, "feature_2", "", "variation_1"),
        ("feature_2", expected_forced_decision_with_rule_key, 200, "feature_2", "feature_2_test", "variation_2")
    ],
    ids=["variation_1", "16931381940"])
def test_decide_with_forced_decision__feature(session_obj, flag_key, expected_response, expected_status_code,
                                              forced_flag, forced_rule, forced_variation):
    """
    Test validates:
    Correct response when valid or empty rule key is passed in forced-decision parameters.
    ...
    :param session_obj:
    :param flag_key:
    :param expected_response:
    :param expected_status_code:
    :param forced_flag:
    :param forced_rule:
    :param forced_variation:
    """

    payload = {
        "userId": "matjaz",
        "decideOptions": [
            "ENABLED_FLAGS_ONLY",
            "INCLUDE_REASONS"
        ],
        "userAttributes": {"attr_1": "hola"},
        "forcedDecisions": [
            {
                "flagKey": forced_flag,
                "ruleKey": f"{forced_rule}",
                "variationKey": forced_variation,
            }
        ]
    }

    params = {"keys": flag_key}
    resp = create_and_validate_request_and_response(ENDPOINT_DECIDE, "post", session_obj, payload=json.dumps(payload),
                                                    params=params)

    assert json.loads(json.dumps(expected_response)) == resp.json()
    assert resp.status_code == expected_status_code, resp.text
    resp.raise_for_status()


def test_decide_403(session_override_sdk_key):
    """
    Test that 403 Forbidden is returned. We use invalid SDK key to trigger 403.
    :param session_override_sdk_key: sdk key to override the session using invalid sdk key
    """
    payload = '{"userId": "matjaz", "userAttributes": {"attr_1": "hola"}}'
    params = {"flagKey": "feature_2"}

    with pytest.raises(requests.exceptions.HTTPError):
        resp = create_and_validate_request_and_response(ENDPOINT_DECIDE, "post", session_override_sdk_key,
                                                        payload=payload, params=params)

        assert resp.status_code == 403
        assert resp.json()["error"] == "unable to fetch fresh datafile (consider " \
                                       "rechecking SDK key), status code: 403 Forbidden"

        resp.raise_for_status()
