import os

import pytest
import requests

from tests.acceptance.helpers import ENDPOINT_OVERRIDE

BASE_URL = os.getenv('host')


# TODO - USE EXISTING EXAMPLE: https://github.com/optimizely/agent/blob/master/examples/override.py
# TODO - only runs when I change to true in config.yaml file

# TODO - NOT WORKING. AFTER FORCING A VARIATION, CONFIG DOESN'T SHOW NEW VARIATION
def test_overrides(session_obj):
    """
    Override an experiment decision for a user

    Override is disabled by default. TO be able to test we enable it in config.yaml in
     enableOverrides: true. TODO - MAKE FUNCTION THAT UPDATES YAML FILE AND THE PUTS BACK
     TODO - TO FALSE - BUT THAT"S DANGEROUS TO MESS UP WITH PUBLIC CODE.
     TODO - IF ERROR CUSTOMER MAY END UP WITH ENABLED OVERRIDE SETTING!
     TODO - WHAT IS SAFER WAY TO DO IT?

    Override a experiment or feature test decision used in future user based decisions.
    This override is only stored locally in memory for debugging and testing purposes
    and should not be used for production overrides.

    Responses from the spec:
        '201':
          description: Forced variation set
        '204':
          description: Forced variation was already set
        '400':
          description: Invalid user id, experiment key, or variation key
        '403':
          $ref: '#/components/responses/Forbidden'

    :param session_obj: session fixture
    """

    # override
    payload = {"userId": "matjaz", "userAttributes": {"attr_1": "hola"},
               "experimentKey": "ab_test1", "variationKey": "my_new_variation"}
    resp = session_obj.post(BASE_URL + ENDPOINT_OVERRIDE, json=payload)

    assert resp.status_code == 201

    # attempt to override again
    resp_second = session_obj.post(BASE_URL + ENDPOINT_OVERRIDE, json=payload)

    assert resp_second.status_code == 204


# TODO - add tests - no deletion, do status code checks, param check,
#  negative status code etc
@pytest.mark.parametrize("userId, experimentKey, variationKey, expected_status_code, expected_error_msg", [
    ("", "ab_test1", "new_variation", 400, '{"error":"userId cannot be empty"}\n'),
    ("matjaz", "", "new_variation", 400, '{"error":"userId cannot be empty"}\n'),
    ("matjaz", "ab_test1", "", 400, '{"error":"userId cannot be empty"}\n'),
], ids=["empty_userId", "empty_experiment_key", "empty_variationKey"])
def test_overrides__invalid_arguments(session_obj, userId, experimentKey, variationKey, expected_status_code, expected_error_msg):
    invalid_user_id = {"userId": "", "userAttributes": {"attr_1": "hola"},
                       "experimentKey": "ab_test1", "variationKey": "my_new_variation"}

    resp = session_obj.post(BASE_URL + ENDPOINT_OVERRIDE, json=invalid_user_id)

    assert resp.status_code == expected_status_code, resp.text
    assert resp.text == expected_error_msg


def test_overrides_403(session_override_sdk_key):
    """
    Test that 403 Forbidden is returned. We use invalid SDK key to trigger 403.
    :param : session_override_sdk_key
    """
    payload = {"userId": "matjaz", "userAttributes": {"attr_1": "hola"},
               "experimentKey": "ab_test1", "variationKey": "my_new_variation"}

    with pytest.raises(requests.exceptions.HTTPError):
        resp = session_override_sdk_key.post(BASE_URL + ENDPOINT_OVERRIDE, json=payload)

        assert resp.status_code == 403
        assert resp.json()['error'] == 'unable to fetch fresh datafile (consider ' \
                                       'rechecking SDK key), status code: 403 Forbidden'

        resp.raise_for_status()
