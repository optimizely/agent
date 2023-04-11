import json

import pytest
import requests

from tests.acceptance.helpers import ENDPOINT_DECIDE
from tests.acceptance.helpers import create_and_validate_request_and_response
from tests.acceptance.helpers import sort_response

expected_fetch_disabled = {
    "variationKey": "off",
    "enabled": False,
    "ruleKey": "default-rollout-52207-23726430538",
    "flagKey": "flag1",
    "userContext": {
        "userId": "matjaz-user-1",
        "attributes": {}
    },
    "reasons": ['an error occurred while evaluating nested tree for audience ID "23783030150"',
                'Audiences for experiment ab_experiment collectively evaluated to false.',
                'User "matjaz-user-1" does not meet conditions to be in experiment "ab_experiment".',
                'Audiences for experiment default-rollout-52207-23726430538 collectively evaluated to true.',
                'User "matjaz-user-1" meets conditions for targeting rule "Everyone Else".']
}

expected_no_segments_fetched = {
    "variationKey": "off",
    "enabled": False,
    "ruleKey": "default-rollout-52207-23726430538",
    "flagKey": "flag1",
    "userContext": {
        "userId": "test_user",
        "attributes": {}
    },
    "reasons": ['an error occurred while evaluating nested tree for audience ID "23783030150"',
                'Audiences for experiment ab_experiment collectively evaluated to false.',
                'User "test_user" does not meet conditions to be in experiment "ab_experiment".',
                'Audiences for experiment default-rollout-52207-23726430538 collectively evaluated to true.',
                'User "test_user" meets conditions for targeting rule "Everyone Else".']
}

expected_fetch_enabled = {
    "variationKey": "variation_b",
    "enabled": True,
    "ruleKey": "ab_experiment",
    "flagKey": "flag1",
    "userContext": {
        "userId": "matjaz-user-1",
        "attributes": {}
    },
    "reasons": ["Audiences for experiment ab_experiment collectively evaluated to true."]
}

expected_fetch_enabled_default_rollout = {
    "variationKey": "off",
    "enabled": False,
    "ruleKey": "default-rollout-52231-23726430538",
    "flagKey": "flag2",
    "userContext": {
        "userId": "matjaz-user-1",
        "attributes": {}
    },
    "reasons": ['Audiences for experiment default-rollout-52231-23726430538 collectively evaluated to true.',
                'User "matjaz-user-1" meets conditions for targeting rule "Everyone Else".']
}

expected_fetch_failed = {
    'error': 'failed to fetch qualified segments'
}


@pytest.mark.parametrize(
    "flag_key, expected_response, expected_status_code, fetch_segments, user_id", [
        ("flag1", expected_fetch_disabled, 200, False, 'matjaz-user-1'),
        ("flag1", expected_fetch_enabled, 200, True, 'matjaz-user-1'),
        ("flag2", expected_fetch_enabled_default_rollout, 200, True, 'matjaz-user-1'),
        ("flag1", expected_no_segments_fetched, 200, True, 'test_user'),
    ])
def test_decide_fetch_qualified_segments(session_override_sdk_key_odp, flag_key, expected_response, expected_status_code, fetch_segments, user_id):
    """
    Test validates:
    Correct response with fetch_segments enabled and disabled.
    ...
    :param session_override_sdk_key_odp: sdk key to override the session using odp key
    :param flag_key:
    :param expected_response:
    :param expected_status_code:
    :param fetch_segments:
    :param user_id:
    """

    payload = {
        "userId": user_id,
        "decideOptions": [
            "ENABLED_FLAGS_ONLY",
            "INCLUDE_REASONS"
        ],
        "userAttributes": {},
        "fetchSegments": fetch_segments,
    }

    params = {"keys": flag_key}
    resp = create_and_validate_request_and_response(ENDPOINT_DECIDE, "post", session_override_sdk_key_odp, payload=json.dumps(payload),
                                                    params=params)

    assert json.loads(json.dumps(expected_response)) == resp.json()
    assert resp.status_code == expected_status_code, resp.text
    resp.raise_for_status()


@pytest.mark.parametrize(
    "flag_key, expected_response, expected_status_code", [
        ("flag1", expected_fetch_failed, 500),
    ])
def test_decide_fetch_qualified_segments_odp_not_integrated(session_obj, flag_key, expected_response, expected_status_code):
    """
    Test validates:
    Correct response when odp not integrated.
    ...
    :param session_override_sdk_key_odp: sdk key to override the session using odp key
    :param flag_key:
    :param expected_response:
    :param expected_status_code:
    """

    payload = {
        "userId": "matjaz-user-1",
        "decideOptions": [
            "ENABLED_FLAGS_ONLY",
            "INCLUDE_REASONS"
        ],
        "userAttributes": {},
        "fetchSegments": True,
    }

    params = {"keys": flag_key}

    with pytest.raises(requests.exceptions.HTTPError):
        resp = create_and_validate_request_and_response(ENDPOINT_DECIDE, "post", session_obj, payload=json.dumps(payload),
                                                        params=params)

        assert json.loads(json.dumps(expected_response)) == resp.json()
        assert resp.status_code == expected_status_code, resp.text
        resp.raise_for_status()
