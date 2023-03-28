import json

import pytest
import requests

from tests.acceptance.helpers import ENDPOINT_SEND_ODP_EVENT
from tests.acceptance.helpers import create_and_validate_request_and_response

expected_odp_not_integrated = {
    'error': 'ODP is not integrated'
}

expected_missing_identifiers = {
    'error': 'missing or empty "identifiers" in request payload'
}

expected_invalid_data = {
    'error': 'ODP data is not valid'
}

expected_missing_action = {
    'error': 'missing "action" in request payload'
}

expected_request_error = {
    'error': 'error parsing request body'
}

expected_successful_odp_event = {
    "success": True
}


@pytest.mark.parametrize(
    "expected_response, expected_status_code", [
        (expected_odp_not_integrated, 500)
    ])
def test_odp_not_integrated(session_obj, expected_response, expected_status_code):
    """
    Test validates:
    Correct response when valid parameters are provided.
    ...
    :param session_obj:
    :param expected_response:
    :param expected_status_code:
    """

    payload = {
        "type": "test_type",
        "data": {
            "idempotence_id": "abc-1234",
            "data_source_type": "agent",
        },
        "identifiers": {"user_id": "test_user_1"},
        "action": "test_action",
    }

    with pytest.raises(requests.exceptions.HTTPError):
        resp = create_and_validate_request_and_response(ENDPOINT_SEND_ODP_EVENT, "post", session_obj, payload=json.dumps(payload),
                                                        params={})

        assert json.loads(json.dumps(expected_response)) == resp.json()
        assert resp.status_code == expected_status_code
        assert resp.text == '{"error":"ODP is not integrated"}\n'
        resp.raise_for_status()


@pytest.mark.parametrize(
    "expected_response, expected_status_code", [
        (expected_successful_odp_event, 200)
    ])
def test_send_odp_event_valid_payload(session_override_sdk_key_odp, expected_response, expected_status_code):
    """
    Test validates:
    Correct response when valid parameters are provided.
    ...
    :param session_override_sdk_key_odp: sdk key to override the session using odp key
    :param expected_response:
    :param expected_status_code:
    """

    payload = {
        "type": "test_type",
        "data": {
            "idempotence_id": "abc-1234",
            "data_source_type": "agent",
        },
        "identifiers": {"user_id": "test_user_1"},
        "action": "test_action",
    }

    resp = create_and_validate_request_and_response(ENDPOINT_SEND_ODP_EVENT, "post", session_override_sdk_key_odp, payload=json.dumps(payload),
                                                    params={})

    assert json.loads(json.dumps(expected_response)) == resp.json()
    assert resp.status_code == expected_status_code, resp.text
    resp.raise_for_status()


@pytest.mark.parametrize(
    "identifiers, action, expected_response, expected_status_code, expected_error, bypass_validation_request, bypass_validation_response", [
        ({}, "test_action", expected_missing_identifiers, 400,
         '{"error":"missing or empty \\"identifiers\\" in request payload"}\n', False, False),
        ("", "test_action", expected_request_error, 400,
         '{"error":"error parsing request body"}\n', True, True),
        ({"user_id": "test_user_1"}, "", expected_missing_action,
         400, '{"error":"missing \\"action\\" in request payload"}\n', False, False)
    ])
def test_send_odp_event_invalid_parameters(session_override_sdk_key_odp, identifiers, action, expected_response, expected_status_code, expected_error, bypass_validation_request,
                                           bypass_validation_response):
    """
    Test validates:
    Returns error when empty identifiers or action is provided.
    ...
    :param session_override_sdk_key_odp: sdk key to override the session using odp key
    :param identifiers:
    :param action:
    :param expected_response:
    :param expected_status_code:
    :param expected_error
    """

    payload = {
        "type": "test_type",
        "data": {
            "idempotence_id": "abc-1234",
            "data_source_type": "agent",
        },
        "identifiers": identifiers,
        "action": action,
    }

    with pytest.raises(requests.exceptions.HTTPError):
        resp = create_and_validate_request_and_response(ENDPOINT_SEND_ODP_EVENT, "post", session_override_sdk_key_odp, bypass_validation_request,
                                                        bypass_validation_response, payload=json.dumps(
                                                            payload),
                                                        params={})

        assert json.loads(json.dumps(expected_response)) == resp.json()
        assert resp.status_code == expected_status_code
        assert resp.text == expected_error
        resp.raise_for_status()


@pytest.mark.parametrize(
    "expected_response, expected_status_code, expected_error, bypass_validation_request, bypass_validation_response", [
        (expected_invalid_data, 500,
         '{"error":"ODP data is not valid"}\n', True, True)
    ])
def test_send_odp_event_invalid_data(session_override_sdk_key_odp, expected_response, expected_status_code, expected_error, bypass_validation_request,
                                     bypass_validation_response):
    """
    Test validates:
    Returns error when invalid attributes are provided under data.
    ...
    :param session_override_sdk_key_odp: sdk key to override the session using odp key
    :param expected_response:
    :param expected_status_code:
    :param expected_error
    """

    payload = {
        "type": "test_type",
        "data": {
            "idempotence_id": {"invalid": "kv"},
            "data_source_type": [],
        },
        "identifiers": {"user_id": "test_user_1"},
        "action": "test_action",
    }

    with pytest.raises(requests.exceptions.HTTPError):
        resp = create_and_validate_request_and_response(ENDPOINT_SEND_ODP_EVENT, "post",
                                                        session_override_sdk_key_odp,
                                                        bypass_validation_request,
                                                        bypass_validation_response,
                                                        payload=json.dumps(
                                                            payload),
                                                        params={})

        assert json.loads(json.dumps(expected_response)) == resp.json()
        assert resp.status_code == expected_status_code
        assert resp.text == expected_error
        resp.raise_for_status()


def test_send_odp_event_403(session_override_sdk_key):
    """
    Test that 403 Forbidden is returned. We use invalid SDK key to trigger 403.
    :param session_override_sdk_key: sdk key to override the session using invalid sdk key
    """
    payload = {
        "type": "test_type",
        "data": {
            "idempotence_id": "abc-1234",
            "data_source_type": "agent",
        },
        "identifiers": {"user_id": "test_user_1"},
        "action": "test_action",
    }

    with pytest.raises(requests.exceptions.HTTPError):
        resp = create_and_validate_request_and_response(ENDPOINT_SEND_ODP_EVENT, 'post', session_override_sdk_key,
                                                        payload=json.dumps(payload), params={})

        assert resp.status_code == 403
        assert resp.json()['error'] == 'unable to fetch fresh datafile (consider ' \
            'rechecking SDK key), status code: 403 Forbidden'

        resp.raise_for_status()
