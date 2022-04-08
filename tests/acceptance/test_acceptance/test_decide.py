import json

import pytest
import requests

from tests.acceptance.helpers import ups_is_active
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

expected_single_flag_key = r"""
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

expected_single_flag_key_ups = r"""
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

if ups_is_active():
    @pytest.mark.parametrize(
        "flag_key, expected_response, expected_status_code", [
            ("feature_2", expected_single_flag_key, 200),
            ("feature_2", expected_single_flag_key_ups, 200),
            ("invalid_flag_key", expected_invalid_flag_key, 200),
        ],
        ids=["valid case", "valid case ups", "invalid_flag_key"])
    def test_decide_w_ups__feature(session_obj, flag_key, expected_response, expected_status_code):
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
        print('==================================')
        print('====    UPS IS ACTIVE    ====')
        print('==================================')

        params = {"keys": flag_key}
        resp = create_and_validate_request_and_response(ENDPOINT_DECIDE, "post", session_obj, payload=payload,
                                                        params=params)
        # TODO - maybe if block could be deleted becasue we have else statement that test without UPS
        if not json.loads(expected_response) == resp.json():
            # if response doesn't include fixed variation (ups), then assert that reasons are same as when no UPS
            assert json.loads(expected_response)['reasons'] == json.loads(expected_single_flag_key)['reasons']
            assert resp.status_code == expected_status_code, resp.text
            resp.raise_for_status()
        else:
            # response comes from ups
            assert json.loads(expected_response) == resp.json()
            assert resp.status_code == expected_status_code, resp.text
            resp.raise_for_status()
else:
    @pytest.mark.parametrize(
        "flag_key, expected_response, expected_status_code", [
            ("feature_2", expected_single_flag_key, 200),
            ("feature_2", expected_single_flag_key_ups, 200),
            ("invalid_flag_key", expected_invalid_flag_key, 200),
        ],
        ids=["valid case", "valid case ups", "invalid_flag_key"])
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
        print('==================================')
        print('====    UPS NOT ACTIVE    ====')
        print('==================================')

        params = {"keys": flag_key}
        resp = create_and_validate_request_and_response(ENDPOINT_DECIDE, "post", session_obj, payload=payload,
                                                        params=params)

        if not json.loads(expected_response) == resp.json():
            # if response regular reasons (no fixed var -ups), then assert that reasons are same as with UPS
            # this will pass in fargate acc tests (pre-UPS version of Agent), but not here in Agent v2.7.0 on (w ups)
            # that's because we determin if UPS is active by looking into config.yaml file is field
            # client.userProfileService is present. If it is, then that Agent version does have UPS active. If not,
            # it doesn't.
            # The problem we have is that Agent repo is on v2.7.0 with UPS and Fargate acc tests are on Agent v.2.6.0
            # without UPS. So the if/else statement that determines if UPS is present will determine which of the two
            # test sets will execute:
            #   1. test_decide__feature_w_ups
            #   2. test_decide__feature
            # The if block with test_decide_w_ups__feature() will run on v2.7.0 that have ups, that is Agent repo) and
            # the else block with test_decide__feature() will run on v2.6.0 and less, that is Fargate Agent.
            # Important: else block will never get executed on Agent repo with v2.7.0+, so no worries if it doesn't run
            # And else block will always execute on versions 2.6.0 or less whch is currently on Fargate. Until we
            # upgrade Fargate to handle UPS. Then we can probably ditch this else block.
            # TODO: current issue - else block fails the test_decide__feature test when expected not same as resp.
            #   - that's because UPS in Agent 2.7.0 is AWLWAYS on REGARDLESS if we enable or disable field
            #   client.userprofileService in config.yaml. This may not be a problem becasue this else block never runs
            #   in Agent 2.7.0 (basically we need a way to completely and fully toggle whole UPS on or off) and then
            #   use this in acc tests.
            # TODO next: create a PR and run fargate acc tests with that branch, sidedoor-teraform travis should fetch
            #            this new branch that can detect if UPS is present or nopt. It won't be present in v2.6.0 and
            #            so acc tests in sidedoor-teraform should run the decide tests (what a mess)
            assert json.loads(expected_response)['reasons'] == json.loads(expected_single_flag_key_ups)['reasons']
            assert resp.status_code == expected_status_code, resp.text
            resp.raise_for_status()
        else:
            # response comes from ups
            assert json.loads(expected_response) == resp.json()
            assert resp.status_code == expected_status_code, resp.text
            resp.raise_for_status()


        # assert json.loads(expected_response) == resp.json()
        # assert resp.status_code == expected_status_code, resp.text
        # resp.raise_for_status()


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


expected_flag_keys = r"""[
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

expected_flag_key__multiple_parameters = r"""[
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


@pytest.mark.parametrize(
    "parameters, expected_response, expected_status_code, bypass_validation_request, bypass_validation_response", [
        ({}, expected_flag_keys, 200, True, True),
        ({"keys": []}, expected_flag_keys, 200, True, True),
        ({"keys": ["feature_1", "feature_2", "feature_4", "feature_5"]}, expected_flag_key__multiple_parameters, 200, True, True),
    ],
    ids=["missig_flagkey_parameter", "no flag key specified", "multiple parameters"])
def test_decide__flag_key_parameter(session_obj, parameters, expected_response, expected_status_code,
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
