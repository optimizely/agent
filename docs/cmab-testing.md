# CMAB Testing Guide for Agent

## Overview

This guide covers testing CMAB (Contextual Multi-Armed Bandit) functionality in Optimizely Agent. CMAB uses machine learning to dynamically optimize which variation a user sees based on their context.

Agent exposes CMAB through the `/v1/decide` API. When a flag uses CMAB, the response includes a `cmabUUID` field that identifies the CMAB decision.

## Setup

### Prerequisites

- Agent binary built and ready
- Optimizely project with a CMAB-enabled flag
- SDK key from your project
- curl command-line tool

### Configure Agent

Edit `config.yaml` to configure CMAB settings:

```yaml
cmab:
  requestTimeout: 10s
  cache:
    type: "memory"
    size: 1000
    ttl: 30m
  retryConfig:
    maxRetries: 3
    initialBackoff: 100ms
    maxBackoff: 10s
    backoffMultiplier: 2.0
```

### Set Environment Variables

```bash
export OPTIMIZELY_SDK_KEY="your_sdk_key_here"
export FLAG_KEY="your_cmab_flag_key"
```

### Start Agent

```bash
cd /Users/matjaz/repositories/agent
./bin/optimizely
```

Agent will start on http://localhost:8080

## Test Cases

### Test 1: Basic CMAB Decision

Make a decide request and verify it returns a cmabUUID.

```bash
curl -X POST "http://localhost:8080/v1/decide?keys=${FLAG_KEY}" \
  -H "X-Optimizely-SDK-Key: ${OPTIMIZELY_SDK_KEY}" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "test_user_1",
    "userAttributes": {
      "age": 25,
      "location": "San Francisco"
    }
  }'
```

Expected response should include:

```json
{
  "flagKey": "your_flag_key",
  "enabled": true,
  "variationKey": "treatment",
  "ruleKey": "cmab_rule",
  "cmabUUID": "550e8400-e29b-41d4-a716-446655440000",
  "variables": { ... }
}
```

What to verify:
- HTTP status is 200
- Response contains cmabUUID field
- UUID is in valid format
- variationKey is returned

### Test 2: Cache Behavior - Same Request Twice

Make the same request twice and verify the second one uses cached results.

First request:
```bash
curl -X POST "http://localhost:8080/v1/decide?keys=${FLAG_KEY}" \
  -H "X-Optimizely-SDK-Key: ${OPTIMIZELY_SDK_KEY}" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user_cache_test",
    "userAttributes": {"age": 30}
  }'
```

Copy the cmabUUID from the response, then make the exact same request again after waiting 1-2 seconds.

What to verify:
- Both requests return the same cmabUUID
- Second request is faster (because it's cached)

If the UUIDs are different, caching isn't working.

### Test 3: Cache Miss When Attributes Change

Verify that changing user attributes triggers a new CMAB decision.

Request with age=25:
```bash
curl -X POST "http://localhost:8080/v1/decide?keys=${FLAG_KEY}" \
  -H "X-Optimizely-SDK-Key: ${OPTIMIZELY_SDK_KEY}" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user_attr_test",
    "userAttributes": {"age": 25, "city": "NYC"}
  }'
```

Note the cmabUUID. Then request with age=35:
```bash
curl -X POST "http://localhost:8080/v1/decide?keys=${FLAG_KEY}" \
  -H "X-Optimizely-SDK-Key: ${OPTIMIZELY_SDK_KEY}" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user_attr_test",
    "userAttributes": {"age": 35, "city": "NYC"}
  }'
```

What to verify:
- UUIDs are different (cache was invalidated)
- Both requests succeed
- Variations might be different based on CMAB model

### Test 4: Different Users Get Different Cache Entries

Request for user1:
```bash
curl -X POST "http://localhost:8080/v1/decide?keys=${FLAG_KEY}" \
  -H "X-Optimizely-SDK-Key: ${OPTIMIZELY_SDK_KEY}" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user_1",
    "userAttributes": {"age": 30}
  }'
```

Request for user2 with same attributes:
```bash
curl -X POST "http://localhost:8080/v1/decide?keys=${FLAG_KEY}" \
  -H "X-Optimizely-SDK-Key: ${OPTIMIZELY_SDK_KEY}" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user_2",
    "userAttributes": {"age": 30}
  }'
```

What to verify:
- Different cmabUUID for each user
- Both requests succeed

### Test 5: Multiple Flags

If you have multiple CMAB-enabled flags, test requesting them together:

```bash
curl -X POST "http://localhost:8080/v1/decide?keys=flag1&keys=flag2" \
  -H "X-Optimizely-SDK-Key: ${OPTIMIZELY_SDK_KEY}" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user_multi",
    "userAttributes": {"age": 28}
  }'
```

Expected response is an array:
```json
[
  {
    "flagKey": "flag1",
    "cmabUUID": "uuid-1",
    ...
  },
  {
    "flagKey": "flag2",
    "cmabUUID": "uuid-2",
    ...
  }
]
```

What to verify:
- Array response with multiple decisions
- Each CMAB flag has its own cmabUUID
- Non-CMAB flags won't have a UUID

### Test 6: Decide Options Work with CMAB

Request with decide options:
```bash
curl -X POST "http://localhost:8080/v1/decide?keys=${FLAG_KEY}" \
  -H "X-Optimizely-SDK-Key: ${OPTIMIZELY_SDK_KEY}" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user_options",
    "userAttributes": {"age": 32},
    "decideOptions": ["INCLUDE_REASONS", "EXCLUDE_VARIABLES"]
  }'
```

What to verify:
- reasons array is populated
- variables field is excluded or empty
- cmabUUID is still present

### Test 7: Cache TTL Expiration

To test this, you need to temporarily change the TTL in config.yaml:

```yaml
cmab:
  cache:
    ttl: 1m  # Short TTL for testing
```

Restart Agent, then:
1. Make a request and note the cmabUUID
2. Wait 65 seconds
3. Make the exact same request again

What to verify:
- Second request returns a different cmabUUID (cache expired)

### Test 8: Reset Endpoint

The reset endpoint clears the client cache. This is useful for test isolation.

Make a decision request:
```bash
curl -X POST "http://localhost:8080/v1/decide?keys=${FLAG_KEY}" \
  -H "X-Optimizely-SDK-Key: ${OPTIMIZELY_SDK_KEY}" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user_reset",
    "userAttributes": {"age": 40}
  }'
```

Note the cmabUUID, then call reset:
```bash
curl -X POST "http://localhost:8080/v1/reset" \
  -H "X-Optimizely-SDK-Key: ${OPTIMIZELY_SDK_KEY}"
```

Should return:
```json
{"result": true}
```

Make the same decision request again.

What to verify:
- Reset returns {"result": true}
- After reset, the same request may get a different cmabUUID

### Test 9: Forced Decisions Bypass CMAB

Request with a forced decision:
```bash
curl -X POST "http://localhost:8080/v1/decide?keys=${FLAG_KEY}" \
  -H "X-Optimizely-SDK-Key: ${OPTIMIZELY_SDK_KEY}" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user_forced",
    "userAttributes": {"age": 50},
    "forcedDecisions": [
      {
        "flagKey": "'${FLAG_KEY}'",
        "variationKey": "control"
      }
    ]
  }'
```

What to verify:
- Returns the forced variationKey ("control")
- CMAB is bypassed (may not have cmabUUID)

### Test 10: Configuration Changes

Test that config changes take effect:

**Test different timeout:**
Edit config.yaml:
```yaml
cmab:
  requestTimeout: 30s
```
Restart Agent and verify requests succeed with longer timeout.

**Test smaller cache:**
Edit config.yaml:
```yaml
cmab:
  cache:
    size: 10
```
Restart Agent, make 15 different requests (different users), then repeat request #1. You might get a different UUID because the cache was full.

**Test retry config:**
Edit config.yaml:
```yaml
cmab:
  retryConfig:
    maxRetries: 5
```
Restart Agent and check logs for retry behavior when the prediction endpoint is slow.

## Configuration Reference

Here's what you can configure in config.yaml:

```yaml
cmab:
  # Timeout for CMAB prediction API calls
  requestTimeout: 10s

  cache:
    # Cache type: "memory" or "redis"
    type: "memory"

    # Maximum cached entries (memory cache only)
    size: 1000

    # How long cache entries live
    ttl: 30m

  retryConfig:
    # How many times to retry failed requests
    maxRetries: 3

    # Starting backoff duration
    initialBackoff: 100ms

    # Maximum backoff duration
    maxBackoff: 10s

    # Backoff multiplier (exponential)
    backoffMultiplier: 2.0
```

Settings you might want to test:

| Setting | Default | Try Testing | What Changes |
|---------|---------|-------------|--------------|
| requestTimeout | 10s | 5s, 30s | Request timeout behavior |
| cache.size | 1000 | 10, 100 | Cache eviction |
| cache.ttl | 30m | 1m, 60m | Cache expiration |
| maxRetries | 3 | 0, 5 | Retry attempts |

## What Good Results Look Like

When CMAB is working correctly:

**Decide responses contain:**
- cmabUUID field with a valid UUID
- variationKey from the CMAB decision
- ruleKey referencing the CMAB rule

**Caching works:**
- Same requests return the same UUID
- Changing attributes gets a new decision
- Different users get separate cache entries

**Configuration is respected:**
- TTL expires cache at the right time
- Timeout prevents hanging
- Retries happen as configured

**Endpoints respond:**
- /v1/decide returns decisions
- /v1/reset clears the cache

## Common Issues

### No cmabUUID in the response

Possible reasons:
- Flag isn't configured for CMAB in your project
- User doesn't qualify for the CMAB rule (check attributes)
- Prediction endpoint isn't accessible

Debug by checking Agent logs:
```bash
tail -f agent.log | grep -i cmab
```

### Cache not working

If every request returns a different UUID:
- Check if TTL is too short
- Verify you're not changing userID or attributes
- Check if cache size is too small

Enable debug logging:
```bash
# In config.yaml
log:
  level: debug

# Restart and check logs
./bin/optimizely 2>&1 | grep cache
```

### Requests timeout

Check:
- Is the prediction endpoint responsive?
- Is requestTimeout too short?

Increase timeout temporarily:
```yaml
cmab:
  requestTimeout: 30s
```

## Test Results Template

Document your test results:

```
CMAB Testing Results
Date: __________
Tester: __________
Agent Version: __________
SDK Key: __________

Test 1: Basic CMAB Decision          [ ] Pass [ ] Fail
Test 2: Cache - Same Request          [ ] Pass [ ] Fail
Test 3: Cache - Attribute Change      [ ] Pass [ ] Fail
Test 4: Cache - Different Users       [ ] Pass [ ] Fail
Test 5: Multiple Flags                [ ] Pass [ ] Fail [ ] N/A
Test 6: Decide Options                [ ] Pass [ ] Fail
Test 7: Cache TTL Expiration          [ ] Pass [ ] Fail
Test 8: Reset Endpoint                [ ] Pass [ ] Fail
Test 9: Forced Decisions              [ ] Pass [ ] Fail
Test 10: Configuration Changes        [ ] Pass [ ] Fail

Overall: [ ] Pass [ ] Fail

Notes:
_______________________________________
```

## Quick Reference

**API Endpoints:**
- POST /v1/decide - Get feature decisions (includes CMAB)
- POST /v1/reset - Clear client cache
- GET /health - Health check

**Required Headers:**
```
X-Optimizely-SDK-Key: your_sdk_key
Content-Type: application/json
```

**Sample Request:**
```json
{
  "userId": "test_user",
  "userAttributes": {
    "age": 25,
    "location": "SF"
  }
}
```

**CMAB Response Fields:**
```json
{
  "flagKey": "string",
  "enabled": true,
  "variationKey": "string",
  "ruleKey": "string",
  "cmabUUID": "550e8400-e29b-41d4-a716-446655440000",
  "variables": {},
  "userContext": {}
}
```
