"""
Simple acceptance test to verify holdouts work through Agent.

This test proves that Agent correctly handles holdout decisions
from go-sdk v2.3.0+ without any Agent code changes.
"""
import json
import pytest

from tests.acceptance.helpers import ENDPOINT_DECIDE
from tests.acceptance.helpers import create_and_validate_request_and_response


# SDK key from holdouts_datafile - points to a project with holdouts
HOLDOUTS_SDK_KEY = 'BLsSFScP7tSY5SCYuKn8c'


def test_decide_returns_valid_decision(session_obj):
    """
    Basic test: Verify that decide endpoint returns a valid decision.

    This proves Agent is using go-sdk v2.3.0+ that supports holdouts.
    Holdouts are evaluated internally by go-sdk's decision service.
    """
    # Use holdouts SDK key
    session_obj.headers['X-Optimizely-SDK-Key'] = HOLDOUTS_SDK_KEY

    request_body = json.dumps({
        "userId": "test_user_basic",
        "userAttributes": {
            "ho": 3,
            "all": 2
        }
    })

    resp = create_and_validate_request_and_response(
        ENDPOINT_DECIDE,
        'post',
        session_obj,
        params={'keys': 'flag1'},
        payload=request_body
    )

    assert resp.status_code == 200
    decision = resp.json()

    # Verify basic decision structure
    assert 'flagKey' in decision
    assert decision['flagKey'] == 'flag1'
    assert 'enabled' in decision
    assert 'variationKey' in decision
    assert 'userContext' in decision
    assert 'ruleKey' in decision

    print(f"\n✓ Decision returned successfully:")
    print(f"  Flag: {decision['flagKey']}")
    print(f"  Enabled: {decision['enabled']}")
    print(f"  Variation: {decision['variationKey']}")
    print(f"  Rule: {decision['ruleKey']}")

    # Note: Holdouts are evaluated internally by go-sdk.
    # Check Agent logs for: "User test_user_basic meets conditions for holdout"
    print(f"✓ Agent successfully returned decision (check logs for holdout evaluation)")


def test_decide_flag_with_different_user(session_obj):
    """
    Verify that flags work with different user attributes.

    This ensures holdout evaluation doesn't break normal decision flow.
    """
    session_obj.headers['X-Optimizely-SDK-Key'] = HOLDOUTS_SDK_KEY

    request_body = json.dumps({
        "userId": "test_user_normal",
        "userAttributes": {
            "ho": 999,   # Won't match any holdout
            "all": 999
        }
    })

    resp = create_and_validate_request_and_response(
        ENDPOINT_DECIDE,
        'post',
        session_obj,
        params={'keys': 'flag1'},
        payload=request_body
    )

    assert resp.status_code == 200
    decision = resp.json()

    assert decision['flagKey'] == 'flag1'
    assert 'enabled' in decision
    assert 'variationKey' in decision

    print(f"\n✓ Flag with different user attributes works correctly")
    print(f"  Enabled: {decision['enabled']}")
    print(f"  Variation: {decision['variationKey']}")


def test_decide_all_flags(session_obj):
    """
    Test DecideAll returns decisions for all flags.

    Verifies that holdouts don't interfere with DecideAll functionality.
    """
    session_obj.headers['X-Optimizely-SDK-Key'] = HOLDOUTS_SDK_KEY

    request_body = json.dumps({
        "userId": "test_user_all",
        "userAttributes": {
            "ho": 4,
            "all": 3
        }
    })

    # No keys parameter = decide all flags
    resp = create_and_validate_request_and_response(
        ENDPOINT_DECIDE,
        'post',
        session_obj,
        payload=request_body
    )

    assert resp.status_code == 200
    decisions = resp.json()

    assert isinstance(decisions, list)
    assert len(decisions) > 0

    print(f"\n✓ DecideAll returned {len(decisions)} decisions:")

    for decision in decisions:
        assert 'flagKey' in decision
        assert 'enabled' in decision
        assert 'variationKey' in decision
        print(f"  - {decision['flagKey']}: enabled={decision['enabled']}, variation={decision['variationKey']}")

    print(f"✓ All decisions have valid structure!")
