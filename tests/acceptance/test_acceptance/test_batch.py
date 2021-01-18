import json

from tests.acceptance.helpers import ENDPOINT_BATCH
from tests.acceptance.helpers import create_and_validate_request_and_response
from tests.acceptance.helpers import sort_response
from tests.acceptance.test_acceptance.conftest import sdk_key


def test_batch_valid_reponse(session_obj):
    # TODO - parameterize to feed in different values in the payload: valid SDK string, invalid sdk string,
    #  empty string, integer, boolean, double
    """
    Happy path with a single operation
    :param agent_server: starts agent server with default config
    :param session_obj: session object
    """
    payload = """{
    "operations": [{
        "body": {
            "status": "subscribed",
            "email_address": "freddie@example.com"
        },
        "method": "GET",
        "operationID": "1",
        "url": "/v1/config",
        "params": {
        },
        "headers": {
            "X-Optimizely-SDK-Key": "%s",
            "X-Request-Id": "matjaz_1"
        }
    }]
    }"""% sdk_key

    resp = create_and_validate_request_and_response(ENDPOINT_BATCH, 'post', session_obj, payload=payload,
                                                    bypass_validation=False)

    actual_response = resp.json()
    assert 200 == resp.status_code
    assert 200 == actual_response['response'][0]['status']
    assert 0 == actual_response['errorCount']
    assert 'matjaz_1' == actual_response['response'][0]['requestID']
    assert '1' == actual_response['response'][0]['operationID']
    assert '/v1/config' == actual_response['response'][0]['url']
    assert 'experimentsMap' in actual_response['response'][0]['body']
    assert 'ab_test1', 'feature_2_test' in actual_response['response'][0]['body']['experimentsMap']
    resp.raise_for_status()


expected_body_of_operationid_2 = """{
    "experimentsMap": {
        "ab_exper": {
            "id": "17273802375",
            "key": "ab_exper",
            "variationsMap": {
                "my_single_variation": {
                    "featureEnabled": false,
                    "id": "17266384371",
                    "key": "my_single_variation",
                    "variablesMap": {}
                }
            }
        }
    },
    "featuresMap": {
        "feature1": {
            "experimentsMap": {},
            "id": "15444990338",
            "key": "feature1",
            "variablesMap": {
                "fff": {
                    "id": "15427520260",
                    "key": "fff",
                    "type": "string",
                    "value": "ss"
                }
            }
        }
    },
    "revision": "18"
}"""


def test_batch_valid_response__multiple_operations(session_obj):
    """
    Verify that operations with different sdk keys can be sent in a batch.
    :param agent_server: starts agent server with default config
    :param session_obj: session object
    """
    payload = """{
    "operations": [
    {
      "body": {
        "status": "subscribed",
        "email_address": "freddie@example.com"
      },
      "method": "GET",
      "operationID": "1",
      "url": "/v1/config",
      "params": {},
      "headers": {
        "X-Optimizely-SDK-Key": "%s",
        "X-Request-Id": "matjaz 1"
      }
    },
    {
      "body": {
        "status": "subscribed",
        "email_address": "freddie@example.com"
      },
      "method": "GET",
      "operationID": "2",
      "url": "/v1/config",
      "params": {},
      "headers": {
        "X-Optimizely-SDK-Key": "TkB2xhu8WEAHa4LphN3xZ2"
      }
    },
    {
      "body": {
        "userId": "user1"
      },
      "method": "POST",
      "operationID": "3",
      "url": "/v1/activate",
      "params": {
        "type": "feature",
        "experimentKey": "ab_test1"
      },
      "headers": {
        "X-Optimizely-SDK-Key": "%s",
        "Content-Type": "application/json"
      }
    }]
    }""" % (sdk_key, sdk_key)

    resp = create_and_validate_request_and_response(ENDPOINT_BATCH, 'post', session_obj, payload=payload,
                                                    bypass_validation=False)

    actual_response = resp.json()

    assert 200 == resp.status_code
    assert 0 == actual_response['errorCount']
    responses = len(actual_response['response'])
    assert 3 == responses
    for operation in range(responses):
        assert 200 == actual_response['response'][operation]['status']

    sorted_actual = sort_response(actual_response['response'], 'operationID', 'status')

    assert json.loads(expected_body_of_operationid_2) == sorted_actual[1]['body']
    assert '/v1/config' == sorted_actual[0]['url']
    assert '/v1/config' == sorted_actual[1]['url']
    assert '/v1/activate' == sorted_actual[2]['url']
    resp.raise_for_status()


def test_batch_400(session_obj):
    """
    Invalid JSON, no SDK key in the operations' header.
    :param agent_server: starts agent server with default config
    :param session_obj: session object
    """
    payload = """{
    "operations": [{
        "body": {
            "status": "subscribed",
            "email_address": "freddie@example.com"
        },
        "method": "GET",
        "operationID": "1",
        "url": "/v1/config",
        "params": {},
        "headers": {
            "X-Optimizely-SDK-Key": "",
            "X-Request-Id": "matjaz_1"
        }
    }]
    }"""

    resp = create_and_validate_request_and_response(ENDPOINT_BATCH, 'post', session_obj, payload=payload,
                                                    bypass_validation=False)

    actual_response = resp.json()
    assert 200 == resp.status_code
    assert 1 == actual_response['errorCount']
    assert 400 == actual_response['response'][0]['status']
    assert 'missing required X-Optimizely-SDK-Key header' == actual_response['response'][0]['body']['error']
    resp.raise_for_status()


def test_batch_422(session_obj):
    """
    Set env variable OPTIMIZELY_SERVER_BATCHREQUESTS_OPERATIONSLIMIT to 3 (already set to 3 for all tests)
    Then send 4 operaions, should fail with code 422
    :param operations_limit: starts agent server with custome set operations limit env var
    :param session_obj: session object
    """
    payload = """{
        "operations": [
        {
          "body": {
            "status": "subscribed",
            "email_address": "freddie@example.com"
          },
          "method": "GET",
          "operationID": "1",
          "url": "/v1/config",
          "params": {},
          "headers": {
            "X-Optimizely-SDK-Key": "%s",
            "X-Request-Id": "matjaz 1"
          }
        },
        {
          "body": {
            "status": "subscribed",
            "email_address": "freddie@example.com"
          },
          "method": "GET",
          "operationID": "2",
          "url": "/v1/config",
          "params": {},
          "headers": {
            "X-Optimizely-SDK-Key": "TkB2xhu8WEAHa4LphN3xZ2"
          }
        },
        {
          "body": {
            "userId": "user1"
          },
          "method": "POST",
          "operationID": "3",
          "url": "/v1/activate",
          "params": {
            "type": "experiment",
            "experimentKey": "ab_test1"
          },
          "headers": {
            "X-Optimizely-SDK-Key": "%s",
            "Content-Type": "application/json"
            }
          },
          {
          "body": {
            "userId": "user2"
          },
          "method": "POST",
          "operationID": "4",
          "url": "/v1/activate",
          "params": {
            "type": "experiment",
            "experimentKey": "ab_test1"
          },
          "headers": {
            "X-Optimizely-SDK-Key": "%s",
            "Content-Type": "application/json"
          }
        }]
        }""" % (sdk_key, sdk_key, sdk_key)

    resp = create_and_validate_request_and_response(ENDPOINT_BATCH, 'post', session_obj, payload=payload,
                                                    bypass_validation=False)

    assert 422 == resp.status_code
    assert resp.json()['error'].startswith('too many operations')
