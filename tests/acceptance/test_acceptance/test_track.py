import json

import os

import pytest
import requests

from tests.acceptance.helpers import ENDPOINT_TRACK

BASE_URL = os.getenv('host')


@pytest.mark.parametrize("event_key, status_code", [
    ("myevent", 204),
    ("", 400),
    ("invalid_event_key", 404)
], ids=["Valid event key", "Empty event key", "Invalid event key"])
def test_track(session_obj, event_key, status_code):
    """
    Track event for the given user.
    Track sends event and user details to Optimizely’s analytics backend
    for the analysis of a feature test or experiment.
    :param session_obj: session fixture
    :param event_key: parameterized param
    :param status_code: parameterized param
    """
    # TODO - ADD EVENT TAGS - AND TEST DIFFERENT SCENARIONS WITH EVENT TAGS
    payload = '{"userId": "matjaz", "userAttributes": {"attr_1": "hola"}, "eventTags": {}}'
    params = {"eventKey": event_key}
    resp = session_obj.post(BASE_URL + ENDPOINT_TRACK, params=params,
                            json=json.loads(payload))
    assert resp.status_code == status_code, f'Status code should be {status_code}. {resp.text}'

    if event_key == "":
        with pytest.raises(requests.exceptions.HTTPError):
            assert resp.status_code == status_code
            assert resp.text == '{"error":"missing required path parameter: eventKey"}\n'
            resp.raise_for_status()

    if event_key == "invalid_event_key":
        with pytest.raises(requests.exceptions.HTTPError):
            assert resp.status_code == status_code
            assert resp.text == '{"error":"event with key \\"invalid_event_key\\" not found"}\n'
            resp.raise_for_status()


def test_track_403(session_override_sdk_key):
    """
    Test that 403 Forbidden is returned. We use invalid SDK key to trigger 403.
    :param : session_obj
    """
    payload = '{"userId": "matjaz", "userAttributes": {"attr_1": "hola"}}'
    params = {"eventKey": "myevent"}

    with pytest.raises(requests.exceptions.HTTPError):
        resp = session_override_sdk_key.post(BASE_URL + ENDPOINT_TRACK, params=params,
                                             json=json.loads(payload))

        assert resp.status_code == 403
        assert resp.json()['error'] == 'unable to fetch fresh datafile (consider ' \
                                       'rechecking SDK key), status code: 403 Forbidden'

        resp.raise_for_status()
