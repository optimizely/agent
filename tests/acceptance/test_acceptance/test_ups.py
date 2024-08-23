import json

from tests.acceptance.helpers import ENDPOINT_DECIDE, ENDPOINT_LOOKUP, ENDPOINT_SAVE
from tests.acceptance.helpers import create_and_validate_request_and_response


def test_ups__feature(session_obj):
    """
    Test validates:
    Same decide call is returned after calling multiple times.
    ...
    :param session_obj:
    """
    payload = """
        {
          "userId": "matjaz",
          "decideOptions": [
              "ENABLED_FLAGS_ONLY"
          ],
          "userAttributes": {"attr_1": "hola"}
        }
    """

    expected_response = """
        {
          "variationKey": "variation_1",
          "enabled": true,
          "isEveryoneElseVariation": false,
          "ruleKey": "feature_2_test",
          "flagKey": "feature_2",
          "userContext": {
            "userId": "matjaz",
            "attributes": {
              "attr_1": "hola"
            }
          },
          "reasons": []
        }
   """

    params = {"keys": 'feature_2'}
    for _ in range(5):
        resp = create_and_validate_request_and_response(ENDPOINT_DECIDE, 'post', session_obj, payload=payload,
                                                        params=params)

        assert json.loads(expected_response) == resp.json()
        assert resp.status_code == 200, resp.text
        resp.raise_for_status()


def test_ups__save(session_obj):
    """
    Test validates:
    Saved variation is returned on lookup and decide calls.
    ...
    :param session_obj:
    """

    save_payload = """
        {
            "experimentBucketMap": {
                "16910084756": {
                    "variation_id": "16915611472"
                }
            },
            "userId": "user1"
        }
    """

    lookup_payload = """
        {
            "userId": "user1"
        }
    """

    lookup_response = """
        {
            "experimentBucketMap": {
                "16910084756": {
                    "variation_id": "16915611472"
                }
            },
            "userId": "user1"
        }
    """

    decide_payload = """
        {
            "userId": "user1",
            "decideOptions": [
                "ENABLED_FLAGS_ONLY"
            ],
            "userAttributes": {"attr_1": "hola"}
        }
    """

    expected_after_saved_variation = """
        {
            "variationKey": "variation_2",
            "enabled": true,
            "ruleKey": "feature_2_test",
            "isEveryoneElseVariation": false,
            "flagKey": "feature_2",
            "userContext": {
                "userId": "user1",
                "attributes": {
                    "attr_1": "hola"
                }
            },
            "reasons": []
        }
    """

    resp = create_and_validate_request_and_response(
        ENDPOINT_SAVE, 'post', session_obj, payload=save_payload)

    assert resp.status_code == 200, resp.text
    resp.raise_for_status()

    resp = create_and_validate_request_and_response(
        ENDPOINT_LOOKUP, 'post', session_obj, payload=lookup_payload)

    assert resp.status_code == 200, resp.text
    resp.raise_for_status()
    assert json.loads(lookup_response) == resp.json()

    params = {"keys": 'feature_2'}
    resp = create_and_validate_request_and_response(ENDPOINT_DECIDE, 'post', session_obj, payload=decide_payload,
                                                    params=params)

    assert json.loads(expected_after_saved_variation) == resp.json()
    assert resp.status_code == 200, resp.text
    resp.raise_for_status()


def test_ups__save_with_invalid_payload(session_obj):
    """
    Test validates:
    No variation is saved with invalid payload.
    ...
    :param session_obj:
    """

    invalid_save_payload = """
        {
            "experimentBucketMaps": {
                "16910084756": {
                    "variation_id": "16915611472"
                }
            },
            "userId": "user1"
        }
    """

    lookup_payload = """
        {
            "userId": "user1"
        }
    """

    lookup_response = """
        {
            "experimentBucketMap": {
            },
            "userId": "user1"
        }
    """

    decide_payload = """
        {
            "userId": "user1",
            "decideOptions": [
                "ENABLED_FLAGS_ONLY"
            ],
            "userAttributes": {"attr_1": "hola"}
        }
    """

    expected_after_saved_variation = """
        {
            "variationKey": "variation_1",
            "enabled": true,
            "ruleKey": "feature_2_test",
            "flagKey": "feature_2",
            "userContext": {
                "userId": "user1",
                "attributes": {
                    "attr_1": "hola"
                }
            },
            "reasons": [],
            "isEveryoneElseVariation": False
        }
    """

    resp = create_and_validate_request_and_response(
        ENDPOINT_SAVE, 'post', session_obj, payload=invalid_save_payload)

    assert resp.status_code == 200, resp.text
    resp.raise_for_status()

    resp = create_and_validate_request_and_response(
        ENDPOINT_LOOKUP, 'post', session_obj, payload=lookup_payload)

    assert resp.status_code == 200, resp.text
    resp.raise_for_status()
    assert json.loads(lookup_response) == resp.json()

    params = {"keys": 'feature_2'}
    resp = create_and_validate_request_and_response(ENDPOINT_DECIDE, 'post', session_obj, payload=decide_payload,
                                                    params=params)

    assert json.loads(expected_after_saved_variation) == resp.json()
    assert resp.status_code == 200, resp.text
    resp.raise_for_status()
