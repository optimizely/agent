import json
from tests.acceptance.helpers import ENDPOINT_DECIDE
from tests.acceptance.helpers import create_and_validate_request_and_response


# CMAB experiment configuration from datafile
CMAB_FLAG_KEY = "cmab_flag"
CMAB_EXPERIMENT_KEY = "cmab-rule_1"


def test_cmab_decision_basic(session_obj):
    """
    Test validates basic CMAB decision flow.

    The test:
    1. Makes a decide request with a user who matches the CMAB audience (attr_1 == "hola")
    2. Verifies the decision returns a variation from the CMAB experiment
    3. Verifies the ruleKey matches the CMAB experiment key

    This test hits the real CMAB prediction endpoint via the agent.
    """
    payload = {
        "userId": "test_user_cmab_1",
        "userAttributes": {"attr_1": "hola"},
        "decideOptions": ["INCLUDE_REASONS"]
    }

    params = {"keys": CMAB_FLAG_KEY}
    resp = create_and_validate_request_and_response(
        ENDPOINT_DECIDE,
        "post",
        session_obj,
        payload=json.dumps(payload),
        params=params
    )

    result = resp.json()

    # Validate response
    assert resp.status_code == 200, f"Expected 200, got {resp.status_code}: {result}"
    assert result["flagKey"] == CMAB_FLAG_KEY
    assert result["ruleKey"] == CMAB_EXPERIMENT_KEY
    assert result["variationKey"] in ["on", "off"], f"Unexpected variation: {result['variationKey']}"
    assert result["userContext"]["userId"] == "test_user_cmab_1"
    assert result["userContext"]["attributes"]["attr_1"] == "hola"


def test_cmab_decision_different_users(session_obj):
    """
    Test validates CMAB decisions for different users.

    Different users with same attributes should potentially get different variations
    based on CMAB ML model predictions.
    """
    users = ["cmab_user_1", "cmab_user_2", "cmab_user_3"]
    variations_seen = set()

    for user_id in users:
        payload = {
            "userId": user_id,
            "userAttributes": {"attr_1": "hola"},
            "decideOptions": []
        }

        params = {"keys": CMAB_FLAG_KEY}
        resp = create_and_validate_request_and_response(
            ENDPOINT_DECIDE,
            "post",
            session_obj,
            payload=json.dumps(payload),
            params=params
        )

        result = resp.json()

        assert resp.status_code == 200
        assert result["flagKey"] == CMAB_FLAG_KEY
        assert result["ruleKey"] == CMAB_EXPERIMENT_KEY
        variations_seen.add(result["variationKey"])

    # Verify we got valid variations
    assert variations_seen.issubset({"on", "off"}), f"Invalid variations: {variations_seen}"


def test_cmab_caching_same_user_same_attributes(session_obj):
    """
    Test validates CMAB caching behavior.

    Same user with identical attributes should get the same variation
    on subsequent requests (cache hit).
    """
    payload = {
        "userId": "test_user_cache",
        "userAttributes": {"attr_1": "hola"},
        "decideOptions": []
    }

    params = {"keys": CMAB_FLAG_KEY}

    # First request
    resp1 = create_and_validate_request_and_response(
        ENDPOINT_DECIDE,
        "post",
        session_obj,
        payload=json.dumps(payload),
        params=params
    )

    result1 = resp1.json()
    assert resp1.status_code == 200
    first_variation = result1["variationKey"]

    # Second request - should return same variation (cached)
    resp2 = create_and_validate_request_and_response(
        ENDPOINT_DECIDE,
        "post",
        session_obj,
        payload=json.dumps(payload),
        params=params
    )

    result2 = resp2.json()
    assert resp2.status_code == 200
    assert result2["variationKey"] == first_variation, "Cache should return same variation"


def test_cmab_cache_invalidation_different_attributes(session_obj):
    """
    Test validates cache invalidation when user attributes change.

    Same user with different attributes may get a different variation
    (cache should be invalidated).
    """
    user_id = "test_user_invalidate"
    params = {"keys": CMAB_FLAG_KEY}

    # First request with attr_1 = "hola"
    payload1 = {
        "userId": user_id,
        "userAttributes": {"attr_1": "hola"},
        "decideOptions": []
    }

    resp1 = create_and_validate_request_and_response(
        ENDPOINT_DECIDE,
        "post",
        session_obj,
        payload=json.dumps(payload1),
        params=params
    )

    assert resp1.status_code == 200
    result1 = resp1.json()
    assert result1["ruleKey"] == CMAB_EXPERIMENT_KEY

    # Second request with same user but additional attribute
    # This should invalidate cache and potentially get different variation
    payload2 = {
        "userId": user_id,
        "userAttributes": {"attr_1": "hola", "extra_attr": "test_value"},
        "decideOptions": []
    }

    resp2 = create_and_validate_request_and_response(
        ENDPOINT_DECIDE,
        "post",
        session_obj,
        payload=json.dumps(payload2),
        params=params
    )

    assert resp2.status_code == 200
    result2 = resp2.json()
    # Both should be CMAB decisions
    assert result2["ruleKey"] == CMAB_EXPERIMENT_KEY


def test_cmab_audience_mismatch(session_obj):
    """
    Test validates CMAB behavior when user doesn't match audience.

    User without matching attributes (attr_1 != "hola") should get
    default rollout decision, not CMAB decision.
    """
    payload = {
        "userId": "test_user_no_match",
        "userAttributes": {"attr_1": "adios"},  # Doesn't match audience requirement
        "decideOptions": ["INCLUDE_REASONS"]
    }

    params = {"keys": CMAB_FLAG_KEY}
    resp = create_and_validate_request_and_response(
        ENDPOINT_DECIDE,
        "post",
        session_obj,
        payload=json.dumps(payload),
        params=params
    )

    result = resp.json()

    # Should fall back to default rollout (off variation)
    assert resp.status_code == 200
    assert result["flagKey"] == CMAB_FLAG_KEY
    assert result["variationKey"] == "off"
    assert result["enabled"] is False
    # Rule key should be the default rollout, not the CMAB experiment
    assert "default-rollout" in result["ruleKey"], f"Expected default rollout, got {result['ruleKey']}"
    assert result["ruleKey"] != CMAB_EXPERIMENT_KEY


def test_cmab_no_attributes(session_obj):
    """
    Test validates CMAB behavior when user has no attributes.

    User without any attributes should not match CMAB audience and
    should get default rollout.
    """
    payload = {
        "userId": "test_user_no_attrs",
        "userAttributes": {},
        "decideOptions": []
    }

    params = {"keys": CMAB_FLAG_KEY}
    resp = create_and_validate_request_and_response(
        ENDPOINT_DECIDE,
        "post",
        session_obj,
        payload=json.dumps(payload),
        params=params
    )

    result = resp.json()

    assert resp.status_code == 200
    assert result["flagKey"] == CMAB_FLAG_KEY
    assert "default-rollout" in result["ruleKey"]


def test_cmab_with_forced_decision(session_obj):
    """
    Test validates forced decisions override CMAB predictions.

    Forced decisions should take precedence over CMAB logic.
    """
    payload = {
        "userId": "test_user_forced",
        "userAttributes": {"attr_1": "hola"},
        "decideOptions": ["INCLUDE_REASONS"],
        "forcedDecisions": [
            {
                "flagKey": CMAB_FLAG_KEY,
                "ruleKey": CMAB_EXPERIMENT_KEY,
                "variationKey": "off"
            }
        ]
    }

    params = {"keys": CMAB_FLAG_KEY}
    resp = create_and_validate_request_and_response(
        ENDPOINT_DECIDE,
        "post",
        session_obj,
        payload=json.dumps(payload),
        params=params
    )

    result = resp.json()

    # Forced decision should override CMAB prediction
    assert resp.status_code == 200
    assert result["variationKey"] == "off"
    assert result["enabled"] is False
    assert "forced decision" in result["reasons"][0].lower()


def test_cmab_decide_all_includes_cmab_flag(session_obj):
    """
    Test validates that DecideAll includes CMAB flag decisions.

    When no specific flag keys are requested, the response should include
    all flags including the CMAB flag.
    """
    payload = {
        "userId": "test_user_decide_all",
        "userAttributes": {"attr_1": "hola"},
        "decideOptions": ["ENABLED_FLAGS_ONLY"]
    }

    # No keys parameter = decide all
    resp = create_and_validate_request_and_response(
        ENDPOINT_DECIDE,
        "post",
        session_obj,
        payload=json.dumps(payload),
        bypass_validation_request=True,
        bypass_validation_response=True
    )

    results = resp.json()

    assert resp.status_code == 200
    assert isinstance(results, list), "DecideAll should return a list"

    # Find the CMAB flag in results
    cmab_result = None
    for result in results:
        if result["flagKey"] == CMAB_FLAG_KEY:
            cmab_result = result
            break

    assert cmab_result is not None, f"CMAB flag '{CMAB_FLAG_KEY}' not found in DecideAll response"
    assert cmab_result["ruleKey"] == CMAB_EXPERIMENT_KEY
    assert cmab_result["variationKey"] in ["on", "off"]


def test_cmab_with_multiple_keys(session_obj):
    """
    Test validates CMAB decision when requested with multiple flag keys.
    """
    payload = {
        "userId": "test_user_multiple",
        "userAttributes": {"attr_1": "hola"},
        "decideOptions": []
    }

    # Request multiple flags including CMAB flag
    params = {"keys": ["feature_1", CMAB_FLAG_KEY, "feature_2"]}
    resp = create_and_validate_request_and_response(
        ENDPOINT_DECIDE,
        "post",
        session_obj,
        payload=json.dumps(payload),
        params=params
    )

    results = resp.json()

    assert resp.status_code == 200
    assert isinstance(results, list)
    assert len(results) >= 3  # At least the 3 requested flags

    # Find CMAB flag result
    cmab_result = None
    for result in results:
        if result["flagKey"] == CMAB_FLAG_KEY:
            cmab_result = result
            break

    assert cmab_result is not None
    assert cmab_result["ruleKey"] == CMAB_EXPERIMENT_KEY
    assert cmab_result["variationKey"] in ["on", "off"]
    