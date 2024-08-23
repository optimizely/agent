import json

import pytest
import requests
import redis
import time

from tests.acceptance.helpers import ENDPOINT_DECIDE
from tests.acceptance.helpers import create_and_validate_request_and_response
from tests.acceptance.helpers import sort_response

expected_redis_save = {
    "variationKey": "variation_b",
    "enabled": True,
    "ruleKey": "ab_experiment",
    "flagKey": "flag1",
    "isEveryoneElseVariation": false,
    "userContext": {
        "userId": "matjaz-user-1",
        "attributes": {}
    },
    "reasons": ['User "matjaz-user-1" was previously bucketed into variation "variation_b" of experiment "ab_experiment".']
}

expected_redis_lookup = {
    "variationKey": "variation_a",
    "enabled": True,
    "ruleKey": "ab_experiment",
    "flagKey": "flag1",
    "isEveryoneElseVariation": false,
    "userContext": {
        "userId": "matjaz-user-2",
        "attributes": {}
    },
    "reasons": ["Audiences for experiment ab_experiment collectively evaluated to true."]
}

expected_redis_reset = {
    "variationKey": "variation_b",
    "enabled": True,
    "ruleKey": "ab_experiment",
    "flagKey": "flag1",
    "isEveryoneElseVariation": false,
    "userContext": {
        "userId": "matjaz-user-4",
        "attributes": {}
    },
    "reasons": ["Audiences for experiment ab_experiment collectively evaluated to true."]
}

def test_redis_save(session_override_sdk_key_odp):
    """
    Test that first fetch call saves segments in redis successfuly.
    :param session_override_sdk_key_odp: sdk key to override the session using odp key
    """
    
    expected_segments = ["atsbugbashsegmenthaspurchased", "atsbugbashsegmentdob"]
    expected_segments_rev = ["atsbugbashsegmentdob", "atsbugbashsegmenthaspurchased"]
    uId = "fs_user_id-$-matjaz-user-1"
    r = redis.Redis(host='localhost', port=6379, db=0)
    # clean redis before testing since several tests use same user_id
    r.flushdb()

    payload = {
        "userId": "matjaz-user-1",
        "decideOptions": [
            "ENABLED_FLAGS_ONLY",
            "INCLUDE_REASONS"
        ],
        "userAttributes": {},
        "fetchSegments": True,
    }

    params = {"keys": "flag1"}
    resp = create_and_validate_request_and_response(ENDPOINT_DECIDE, "post", session_override_sdk_key_odp, payload=json.dumps(payload),
                                                    params=params)

    # Check saved segments
    assert json.loads(json.dumps(expected_segments)) == json.loads(r.get(uId)) or json.loads(json.dumps(expected_segments_rev)) == json.loads(r.get(uId))
        
    assert json.loads(json.dumps(expected_redis_save)) == resp.json()
    assert resp.status_code == 200, resp.text
    resp.raise_for_status()


def test_redis_lookup(session_override_sdk_key_odp):
    """
    Test that saved segments in redis are used successfuly by segment manager using lookup.
    :param session_override_sdk_key_odp: sdk key to override the session using odp key
    """
    
    expected_segments = ["atsbugbashsegmenthaspurchased", "atsbugbashsegmentdob"]
    uId = "fs_user_id-$-matjaz-user-2"
    r = redis.Redis(host='localhost', port=6379, db=0)
    r.set(uId, json.dumps(expected_segments))

    payload = {
        "userId": "matjaz-user-2",
        "decideOptions": [
            "ENABLED_FLAGS_ONLY",
            "INCLUDE_REASONS"
        ],
        "userAttributes": {},
        "fetchSegments": True,
    }

    params = {"keys": "flag1"}
    resp = create_and_validate_request_and_response(ENDPOINT_DECIDE, "post", session_override_sdk_key_odp, payload=json.dumps(payload),
                                                    params=params)

    # Check saved segments
    assert json.loads(json.dumps(expected_segments)) == json.loads(r.get(uId))
        
    assert json.loads(json.dumps(expected_redis_lookup)) == resp.json()
    assert resp.status_code == 200, resp.text
    resp.raise_for_status()


def test_redis_reset(session_override_sdk_key_odp):
    """
    Test that reset option resets segments in redis.
    :param session_override_sdk_key_odp: sdk key to override the session using odp key
    """
    
    expected_segments_after_reset = ['atsbugbashsegmentdob', 'atsbugbashsegmentgender']
    uId = "fs_user_id-$-matjaz-user-4"
    r = redis.Redis(host='localhost', port=6379, db=0)

    # Manually add invalid segments for matjaz-user-4 in redis
    r.set(uId, json.dumps(["invalid_segments"]))
    # Check invalid segments were saved
    assert json.loads(json.dumps(["invalid_segments"])) == json.loads(r.get(uId))

    payload = {
        "userId": "matjaz-user-4",
        "decideOptions": [
            "ENABLED_FLAGS_ONLY",
            "INCLUDE_REASONS"
        ],
        "userAttributes": {},
        "fetchSegments": True,
        "fetchSegmentsOptions": ["RESET_CACHE"],
    }

    params = {"keys": "flag1"}
    resp = create_and_validate_request_and_response(ENDPOINT_DECIDE, "post", session_override_sdk_key_odp, payload=json.dumps(payload),
                                                    params=params)

    # Check old segments were removed and newly fetched were saved
    assert json.loads(json.dumps(expected_segments_after_reset)) == json.loads(r.get(uId))
        
    assert json.loads(json.dumps(expected_redis_reset)) == resp.json()
    assert resp.status_code == 200, resp.text
    resp.raise_for_status()
