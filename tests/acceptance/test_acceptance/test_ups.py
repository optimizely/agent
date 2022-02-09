import json

import pytest

from tests.acceptance.helpers import ENDPOINT_DECIDE, ENDPOINT_LOOKUP, ENDPOINT_SAVE
from tests.acceptance.helpers import create_and_validate_request_and_response


expected_flag_key = """
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
      "reasons": []
    }
"""

expected_after_saved_variation = """
    {
      "variationKey": "variation_2",
      "enabled": true,
      "ruleKey": "feature_2_test",
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


@pytest.mark.parametrize(
    "flag_key, expected_response, expected_status_code", [
        ("feature_2", expected_flag_key, 200)
    ],
    ids=["valid case"])
def test_ups__feature(session_obj, flag_key, expected_response, expected_status_code):
    """
    Test validates:
    Same decide is returned after calling multiple times.
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
              "ENABLED_FLAGS_ONLY"
          ],
          "userAttributes": {"attr_1": "hola"}
        }
    """

    params = {"keys": flag_key}
    for _ in range(5):
      resp = create_and_validate_request_and_response(ENDPOINT_DECIDE, 'post', session_obj, payload=payload,
                                                      params=params)

      assert json.loads(expected_response) == resp.json()
      assert resp.status_code == expected_status_code, resp.text
      resp.raise_for_status()


@pytest.mark.parametrize(
    "expected_status_code", [
        (200)
    ],
    ids=["valid case"])
def test_ups__save(session_obj, expected_status_code):
    """
    Test validates:
    Saved variation is returned on lookup and decide calls.
    ...
    :param session_obj:
    :param expected_status_code:
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

    resp = create_and_validate_request_and_response(
        ENDPOINT_SAVE, 'post', session_obj, payload=save_payload)

    assert resp.status_code == expected_status_code, resp.text
    resp.raise_for_status()

    resp = create_and_validate_request_and_response(
        ENDPOINT_LOOKUP, 'post', session_obj, payload=lookup_payload)

    assert resp.status_code == expected_status_code, resp.text
    resp.raise_for_status()
    assert json.loads(lookup_response) == resp.json()

    params = {"keys": 'feature_2'}
    resp = create_and_validate_request_and_response(ENDPOINT_DECIDE, 'post', session_obj, payload=decide_payload,
                                                    params=params)

    assert json.loads(expected_after_saved_variation) == resp.json()
    assert resp.status_code == expected_status_code, resp.text
    resp.raise_for_status()
