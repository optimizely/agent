import pytest

from tests.acceptance.helpers import ENDPOINT_BATCH
from tests.acceptance.helpers import create_and_validate_request_and_response
from tests.acceptance.helpers import get_pretty_json
from tests.acceptance.test_acceptance.conftest import sdk_key


# TODO Should I parameterize the tests?

# DONE - CLEAN UP !
def test_batch_valid_reponse(session_obj):
    # TODO - parameterize to feed in different values for SDK key: valid SDK string, invalid sdk string, empty string, integer, boolean, double
    """
    Happy path with a single operation
    :param session_obj: session object
    :return: Response with one operation
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
    }""" % sdk_key

    params = {"experimentKey": "ab_test_1"}

    resp = create_and_validate_request_and_response(ENDPOINT_BATCH, 'post', session_obj, payload=payload,
                                                    params=params, bypass_validation=False)

    print(get_pretty_json(resp.json()))

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


# TODO
def test_xxxbatch_valid_response__multiple_operations(session_obj):       # TODO: ADD TWO MORE OPERATIONS, ONE OF THEM W DIFFERENT SDK KEY (HOW TO DO THAT????, NEW PROJECT in QA UI??)
    """
    Verify that operations with different sdk keys can be sent in a batch.
    :param session_obj: session object
    :return:
    """

    payload = """{
    "operations": [
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
    }]
    }""" % sdk_key

    # params = {"experimentKey": "ab_test1"}

    print('>>>>> ', payload)

    resp = create_and_validate_request_and_response(ENDPOINT_BATCH, 'post', session_obj, payload=payload,
                                                    bypass_validation=False)

    print(get_pretty_json(resp.json()))

    actual_response = resp.json()
    assert 200 == resp.status_code
    # assert 200 == actual_response['response'][0]['status']
    assert 0 == actual_response['errorCount']
    # TODO - add assertions on second and third other operation
    resp.raise_for_status()


# DONE - CLEAN UP!
def test_batch_400(session_obj):
    """
    Invalid JSON, no SDK key in the operations' header.
    :param session_obj: session object
    :return: valid response with its response status code 400
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

    params = {"experimentKey": "ab_test_1"}

    resp = create_and_validate_request_and_response(ENDPOINT_BATCH, 'post', session_obj, payload=payload,
                                                    params=params, bypass_validation=False)

    actual_response = resp.json()
    assert 200 == resp.status_code
    assert 1 == actual_response['errorCount']
    assert 400 == actual_response['response'][0]['status']
    assert 'missing required X-Optimizely-SDK-Key header' == actual_response['response'][0]['body']['error']
    resp.raise_for_status()


# TODO
def test_batch_422(session_obj):
    """
    set env variable OPTIMIZELY_SERVER_BATCHREQUESTS_OPERATIONS LIMIT to 4.
        Then send 3 operations -> should pass 200
        Then send 4 operations -> should pass 200
        Then send 5 operaions, should fail with code 422
    :param session_obj: session object
    :return:
    """
    pass
