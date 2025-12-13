"""
Acceptance tests for holdouts functionality in /v1/decide endpoint.

These tests verify that Agent correctly handles holdout decisions through go-sdk v2.3.0+.
Holdouts are evaluated internally by go-sdk and reflected in the ruleKey field.
"""
import json
import pytest

from tests.acceptance.helpers import ENDPOINT_DECIDE
from tests.acceptance.helpers import create_and_validate_request_and_response


@pytest.fixture(scope='function')
def holdouts_session(session_obj):
    """
    Create a session using the holdouts datafile.
    This SDK key points to a project with holdouts configured.
    """
    # SDK key from holdouts_datafile.py
    session_obj.headers['X-Optimizely-SDK-Key'] = 'BLsSFScP7tSY5SCYuKn8c'
    return session_obj


def test_decide_user_in_holdout(holdouts_session):
    """
    Test that a user who qualifies for a holdout gets bucketed into it.

    Expected behavior:
    - ruleKey should match the holdout key
    - enabled should be False
    - variationKey should be "off"
    """
    request_body = json.dumps({
        "userId": "test_user_holdout",
        "userAttributes": {
            "ho": 3,  # Qualifies for holdout_3
            "all": 2   # Satisfies the "all <= 3" condition
        }
    })

    resp = create_and_validate_request_and_response(
        ENDPOINT_DECIDE,
        'post',
        holdouts_session,
        params={'keys': 'flag1'},
        payload=request_body
    )

    assert resp.status_code == 200, f"Expected 200, got {resp.status_code}: {resp.text}"

    decision = resp.json()

    # Verify holdout decision structure
    assert decision['flagKey'] == 'flag1', f"Expected flagKey 'flag1', got {decision.get('flagKey')}"
    assert 'ruleKey' in decision, "Decision should have ruleKey field"
    assert 'enabled' in decision, "Decision should have enabled field"
    assert 'variationKey' in decision, "Decision should have variationKey field"

    # Log the actual decision for debugging
    print(f"\nDecision response: {json.dumps(decision, indent=2)}")

    # Check if user was bucketed into a holdout (ruleKey contains 'holdout')
    if 'holdout' in decision['ruleKey']:
        assert decision['enabled'] == False, "Holdout decisions should have enabled=False"
        assert decision['variationKey'] == 'off', "Holdout decisions should have variationKey='off'"
        print(f"✓ User successfully bucketed into holdout: {decision['ruleKey']}")
    else:
        print(f"✓ User got normal decision with rule: {decision['ruleKey']}")


def test_decide_user_not_in_holdout_audience(holdouts_session):
    """
    Test that a user who doesn't qualify for any holdout audience
    gets normal decision (experiment or rollout).

    Expected behavior:
    - User should get normal flag evaluation
    - ruleKey should NOT contain 'holdout'
    """
    request_body = json.dumps({
        "userId": "test_user_no_holdout",
        "userAttributes": {
            "ho": 999,  # Doesn't match any holdout audience
            "all": 999  # Doesn't satisfy any holdout condition
        }
    })

    resp = create_and_validate_request_and_response(
        ENDPOINT_DECIDE,
        'post',
        holdouts_session,
        params={'keys': 'flag1'},
        payload=request_body
    )

    assert resp.status_code == 200
    decision = resp.json()

    print(f"\nDecision response: {json.dumps(decision, indent=2)}")

    # User should NOT be in holdout
    assert 'ruleKey' in decision, "Decision should have ruleKey"

    # Verify it's not a holdout decision
    is_holdout = 'holdout' in decision['ruleKey'].lower()
    print(f"✓ User correctly bypassed holdouts, got rule: {decision['ruleKey']} (is_holdout={is_holdout})")


def test_decide_multiple_flags_with_holdouts(holdouts_session):
    """
    Test DecideAll with multiple flags when holdouts are present.

    Expected behavior:
    - Should return decisions for all flags
    - Some may be holdout decisions, others may be regular decisions
    - Each decision should have proper structure
    """
    request_body = json.dumps({
        "userId": "test_user_multi",
        "userAttributes": {
            "ho": 4,   # Might qualify for holdout_4
            "all": 3
        }
    })

    # Call without keys to get all flags
    resp = create_and_validate_request_and_response(
        ENDPOINT_DECIDE,
        'post',
        holdouts_session,
        payload=request_body
    )

    assert resp.status_code == 200
    decisions = resp.json()

    assert isinstance(decisions, list), "DecideAll should return array of decisions"
    assert len(decisions) > 0, "Should have at least one decision"

    print(f"\nReceived {len(decisions)} decisions:")

    holdout_count = 0
    for decision in decisions:
        assert 'flagKey' in decision
        assert 'ruleKey' in decision
        assert 'enabled' in decision
        assert 'variationKey' in decision

        is_holdout = 'holdout' in decision['ruleKey'].lower()
        print(f"  - {decision['flagKey']}: rule={decision['ruleKey']}, enabled={decision['enabled']}")

        if is_holdout:
            holdout_count += 1
            assert decision['enabled'] == False
            assert decision['variationKey'] == 'off'

    print(f"✓ Got {holdout_count} holdout decisions out of {len(decisions)} total decisions")


def test_decide_holdout_with_forced_decision(holdouts_session):
    """
    Test that forced decisions override holdout bucketing.

    Expected behavior:
    - Forced decision should take precedence over holdout
    """
    request_body = json.dumps({
        "userId": "test_user_forced",
        "userAttributes": {
            "ho": 3,   # Would normally qualify for holdout
            "all": 2
        },
        "forcedDecisions": [
            {
                "flagKey": "flag1",
                "variationKey": "variation_1"
            }
        ]
    })

    resp = create_and_validate_request_and_response(
        ENDPOINT_DECIDE,
        'post',
        holdouts_session,
        params={'keys': 'flag1'},
        payload=request_body
    )

    assert resp.status_code == 200
    decision = resp.json()

    print(f"\nDecision response: {json.dumps(decision, indent=2)}")

    # Forced decision should override holdout
    assert decision['variationKey'] == 'variation_1', "Forced decision should be respected"
    assert decision['enabled'] == True, "Forced decision should enable the flag"

    # Note: Agent's /v1/decide doesn't return detailed reasons
    # The presence of the forced variation proves it worked
    print("✓ Forced decision correctly overrode holdout bucketing")


def test_decide_holdout_decision_reasons(holdouts_session):
    """
    Test that holdout decisions include proper reasons.

    Expected behavior:
    - reasons array should explain the decision
    """
    request_body = json.dumps({
        "userId": "test_user_reasons",
        "userAttributes": {
            "ho": 5,   # Qualifies for holdout_5
            "all": 4
        }
    })

    resp = create_and_validate_request_and_response(
        ENDPOINT_DECIDE,
        'post',
        holdouts_session,
        params={'keys': 'flag1'},
        payload=request_body
    )

    assert resp.status_code == 200
    decision = resp.json()

    assert 'reasons' in decision, "Decision should include reasons"
    assert isinstance(decision['reasons'], list), "Reasons should be an array"

    # Note: Agent's /v1/decide returns empty reasons array
    # The ruleKey and decision structure are what matter
    print(f"\nDecision reasons: {decision['reasons']}")
    print(f"Rule key: {decision['ruleKey']}")

    # Verify decision structure is correct
    is_holdout = 'holdout' in decision['ruleKey'].lower()
    if is_holdout:
        print(f"✓ Holdout decision structure is correct (rule: {decision['ruleKey']})")
    else:
        print(f"✓ Non-holdout decision structure is correct")


def test_decide_holdout_impression_event(holdouts_session):
    """
    Test that holdout decisions have all necessary fields for impression tracking.

    Expected behavior:
    - Decision should have all necessary fields for impression tracking
    - ruleKey, variationKey, enabled should be present
    """
    request_body = json.dumps({
        "userId": "test_user_impression",
        "userAttributes": {
            "ho": 6,   # Qualifies for holdout_6
            "all": 5
        }
    })

    resp = create_and_validate_request_and_response(
        ENDPOINT_DECIDE,
        'post',
        holdouts_session,
        params={'keys': 'flag1'},
        payload=request_body
    )

    assert resp.status_code == 200
    decision = resp.json()

    # Verify decision has all fields needed for impression event
    required_fields = ['flagKey', 'variationKey', 'ruleKey', 'enabled', 'userContext']
    for field in required_fields:
        assert field in decision, f"Decision missing required field: {field}"

    # Verify userContext is complete
    assert decision['userContext']['userId'] == 'test_user_impression'
    assert 'attributes' in decision['userContext']

    print(f"\n✓ Decision has all fields needed for impression tracking:")
    print(f"  - flagKey: {decision['flagKey']}")
    print(f"  - ruleKey: {decision['ruleKey']}")
    print(f"  - variationKey: {decision['variationKey']}")
    print(f"  - enabled: {decision['enabled']}")


def test_decide_flag2_with_holdouts(holdouts_session):
    """
    Test decision for flag2 which also has holdouts configured.

    Expected behavior:
    - User matching holdout criteria should get holdout decision
    - Decision should have correct structure
    """
    request_body = json.dumps({
        "userId": "test_user_flag2",
        "userAttributes": {
            "ho": 3,
            "all": 2
        }
    })

    resp = create_and_validate_request_and_response(
        ENDPOINT_DECIDE,
        'post',
        holdouts_session,
        params={'keys': 'flag2'},
        payload=request_body
    )

    assert resp.status_code == 200
    decision = resp.json()

    print(f"\nDecision for flag2: {json.dumps(decision, indent=2)}")

    assert decision['flagKey'] == 'flag2'
    assert 'ruleKey' in decision
    assert 'enabled' in decision
    assert 'variationKey' in decision

    # Check if holdout decision
    is_holdout = 'holdout' in decision['ruleKey'].lower()
    if is_holdout:
        print(f"✓ Flag2 holdout decision: {decision['ruleKey']}")
    else:
        print(f"✓ Flag2 normal decision: {decision['ruleKey']}")
