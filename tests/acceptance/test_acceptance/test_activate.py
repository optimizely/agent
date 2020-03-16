import json
import os

import pytest
import requests
from tests.acceptance.helpers import ENDPOINT_ACTIVATE
from tests.acceptance.helpers import ENDPOINT_CONFIG
from tests.acceptance.helpers import sort_response

BASE_URL = os.getenv('host')

expected_activate_ab = """[
    {
        "experimentKey": "ab_test1",
        "featureKey": "",
        "variationKey": "variation_1",
        "type": "experiment",
        "enabled": true
    }
]"""


@pytest.mark.parametrize("experiment_key, expected_response, expected_status_code", [
    ("ab_test1", expected_activate_ab, 200),
    ("", '{"error": "experimentKey not-found"}', 404),
    ("invalid exper key", '{"error": "experimentKey not-found"}', 404),
], ids=["valid case", "empty exper key", "invalid exper key"])
def test_activate__experiment(session_obj, experiment_key, expected_response,
                              expected_status_code):
    """
    Test validates:
    1. Presence of correct variation in the returned decision for AB experiment
    Instead of on single field (variation, enabled), validation is done on the whole
    response (that includes variations and enabled fields).
    This is to add extra robustness to the test.

    Sort the reponses because dictionaries shuffle order.
    :param session_obj: session object
    :param experiment_key: experiment_key
    :param expected_response: expected_response
    :param expected_status_code: expected_status_code
    """
    payload = '{"userId": "matjaz", "userAttributes": {"attr_1": "hola"}}'
    params = {"experimentKey": experiment_key}

    resp = session_obj.post(BASE_URL + ENDPOINT_ACTIVATE, params=params,
                            json=json.loads(payload))

    print('RESP ', resp.json())

    if isinstance(resp.json(), dict) and resp.json()['error']:
        with pytest.raises(requests.exceptions.HTTPError):
            assert resp.json() == json.loads(expected_response)
            assert resp.status_code == expected_status_code, resp.text
            resp.raise_for_status()

    assert json.loads(expected_response) == resp.json()
    assert resp.status_code == expected_status_code, resp.text


expected_activate_feat = """[
  {
    "experimentKey": "",
    "featureKey": "feature_1",
    "variationKey": "",
    "type": "feature",
    "variables": {
      "bool_var": true,
      "double_var": 5.6,
      "int_var": 1,
      "str_var": "hello"
    },
    "enabled": true
  }
]"""


@pytest.mark.parametrize("feature_key, expected_response, expected_status_code", [
    ("feature_1", expected_activate_feat, 200),
    pytest.param("", '{"error": "featureKey not-found"}', 404),
    pytest.param("invalid feat key", '{"error": "featureKey not-found"}', 404),
], ids=["valid case", "empty feat key", "invalid feat key"])
def test_activate__feature(session_obj, feature_key, expected_response,
                           expected_status_code):
    """
    Test validates:
    That feature is enabled in the decision for the feature test
    Instead of on single field (variation, enabled), validation is done on the whole
    response (that includes variations and enabled fields).
    This is to add extra robustness to the test.

    Sort the reponses because dictionaries shuffle order.
    :param session_obj: session object
    """
    payload = '{"userId": "matjaz", "userAttributes": {"attr_1": "hola"}}'
    params = {"featureKey": feature_key}

    resp = session_obj.post(BASE_URL + ENDPOINT_ACTIVATE, params=params,
                            json=json.loads(payload))

    if isinstance(resp.json(), dict) and resp.json()['error']:
        with pytest.raises(requests.exceptions.HTTPError):
            assert resp.json() == json.loads(expected_response)
            assert resp.status_code == expected_status_code, resp.text
            resp.raise_for_status()

    assert json.loads(expected_response) == resp.json()
    assert resp.status_code == expected_status_code, resp.text


expected_activate_type_exper = """[
  {
    "experimentKey": "feature_2_test",
    "featureKey": "",
    "variationKey": "variation_1",
    "type": "experiment",
    "enabled": true
  },
  {
    "experimentKey": "ab_test1",
    "featureKey": "",
    "variationKey": "variation_1",
    "type": "experiment",
    "enabled": true
  }
]"""

expected_activate_type_feat = """[
  {
    "experimentKey": "",
    "featureKey": "feature_2",
    "variationKey": "",
    "type": "feature",
    "enabled": true
  },
  {
    "experimentKey": "",
    "featureKey": "feature_3",
    "variationKey": "",
    "type": "feature",
    "enabled": false
  },
  {
    "experimentKey": "",
    "featureKey": "feature_4",
    "variationKey": "",
    "type": "feature",
    "enabled": true
  },
  {
    "experimentKey": "",
    "featureKey": "feature_5",
    "variationKey": "",
    "type": "feature",
    "enabled": true
  },
  {
    "experimentKey": "",
    "featureKey": "feature_1",
    "variationKey": "",
    "type": "feature",
    "variables": {
      "bool_var": true,
      "double_var": 5.6,
      "int_var": 1,
      "str_var": "hello"
    },
    "enabled": true
  }
]"""


@pytest.mark.parametrize("decision_type, expected_response, expected_status_code", [
    ("experiment", expected_activate_type_exper, 200),
    ("feature", expected_activate_type_feat, 200),
    ("invalid decision type", {'error': 'type "invalid decision type" not supported'},
     400),
    ("", {'error': 'type "" not supported'}, 400)
], ids=["experiment decision type", "feature decision type", "invalid decision type",
        "empty decision type"])
def test_activate__type(session_obj, decision_type, expected_response,
                        expected_status_code):
    """
    Test cases:
    1. Get decisions with "experiment" type
    2. Get decisions with "feature" type
    3. Get empty list when non-existent decision type -> bug OASIS-6031
    :param session_obj: session object
    :param decision_type: parameterized decision type
    :param expected_response: expected response
    """
    # payload = {"userId": "matjaz", "userAttributes": {"attr_1": "hola"}}
    payload = '{"userId": "matjaz", "userAttributes": {"attr_1": "hola"}}'
    params = {"type": decision_type}

    # resp = session_obj.post(BASE_URL + ENDPOINT_ACTIVATE, params=params, json=payload)
    resp = session_obj.post(BASE_URL + ENDPOINT_ACTIVATE, params=params,
                            json=json.loads(payload))

    if decision_type in ['experiment', 'feature']:
        sorted_actual = sort_response(resp.json(), 'experimentKey', 'featureKey')
        sorted_expected = sort_response(json.loads(expected_response), 'experimentKey',
                                        'featureKey')
        assert sorted_actual == sorted_expected
    elif resp.json()['error']:
        with pytest.raises(requests.exceptions.HTTPError):
            assert resp.json() == expected_response
            resp.raise_for_status()


def test_activate_403(session_override_sdk_key):
    """
    Test that 403 Forbidden is returned. We use invalid SDK key to trigger 403.
    :param : session_obj
    """
    payload = '{"userId": "matjaz", "userAttributes": {"attr_1": "hola"}}'
    params = {"type": "experiment"}

    with pytest.raises(requests.exceptions.HTTPError):
        resp = session_override_sdk_key.post(BASE_URL + ENDPOINT_ACTIVATE, params=params,
                                             json=json.loads(payload))

        assert resp.status_code == 403
        assert resp.json()['error'] == 'unable to fetch fresh datafile (consider ' \
                                       'rechecking SDK key), status code: 403 Forbidden'

        resp.raise_for_status()


@pytest.mark.parametrize("experiment, disableTracking, expected_status_code", [
    ("ab_test1", "true", 200),
    ("ab_test1", "false", 200),
    ("feature_2_test", "true", 200),
    ("feature_2_test", "false", 200),
    ("ab_test1", "", 200),
    ("ab_test1", "invalid_boolean", 200),
], ids=["ab_experiment and decision_tr true", "ab_experiment and decision_tr false",
        "feature test and decision_tr true",
        "feature test and decision_tr false", "empty disableTracking",
        "invalid disableTracking"])
def test_activate__disable_tracking(session_obj, experiment, disableTracking,
                                    expected_status_code):
    """
    Setting to true will disable impression tracking for ab experiments and feature tests.
    It's equivalent to previous "get_variation".
    Can not test it in acceptance tests. Just testing basic status code.
    FS compatibility test suite uses proxy event displatcher where they test this by
    validating that event was not sent.
    :param session_obj: session fixture
    :param experiment: ab experiment or feature test
    :param disableTracking: true or false
    :param expected_status_code
    """
    payload = '{"userId": "matjaz", "userAttributes": {"attr_1": "hola"}}'
    params = {
        "experimentKey": experiment,
        "disableTracking": disableTracking
    }

    resp = session_obj.post(BASE_URL + ENDPOINT_ACTIVATE, params=params,
                            json=json.loads(payload))

    resp.raise_for_status()
    assert resp.status_code == expected_status_code


expected_enabled_true_all_true = """[
  {
    "experimentKey": "",
    "featureKey": "feature_1",
    "variationKey": "",
    "type": "feature",
    "variables": {
      "bool_var": true,
      "double_var": 5.6,
      "int_var": 1,
      "str_var": "hello"
    },
    "enabled": true
  },
  {
    "experimentKey": "ab_test1",
    "featureKey": "",
    "variationKey": "variation_1",
    "type": "experiment",
    "enabled": true
  }
]"""

expected_enabled_true_feature_off = """[
  {
    "experimentKey": "ab_test1",
    "featureKey": "",
    "variationKey": "variation_1",
    "type": "experiment",
    "enabled": true
  }
]"""

expected_enabled_false_feature_on = """[]"""

expected_enabled_false_feature_off = """[
  {
    "experimentKey": "",
    "featureKey": "feature_3",
    "variationKey": "",
    "type": "feature",
    "enabled": false
  }
]"""

expected_enabled_empty = """[
  {
    "experimentKey": "ab_test1",
    "featureKey": "",
    "variationKey": "variation_1",
    "type": "experiment",
    "enabled": true
  },
  {
    "experimentKey": "",
    "featureKey": "feature_1",
    "variationKey": "",
    "type": "feature",
    "variables": {
      "bool_var": true,
      "double_var": 5.6,
      "int_var": 1,
      "str_var": "hello"
    },
    "enabled": true
  }
]"""

expected_enabled_invalid = """[
  {
    "experimentKey": "ab_test1",
    "featureKey": "",
    "variationKey": "variation_1",
    "type": "experiment",
    "enabled": true
  },
  {
    "experimentKey": "",
    "featureKey": "feature_1",
    "variationKey": "",
    "type": "feature",
    "variables": {
      "bool_var": true,
      "double_var": 5.6,
      "int_var": 1,
      "str_var": "hello"
    },
    "enabled": true
  }
]"""


@pytest.mark.parametrize(
    "enabled, experimentKey, featureKey, expected_response, expected_status_code", [
        ("true", "ab_test1", "feature_1", expected_enabled_true_all_true, 200),
        ("true", "ab_test1", "feature_3", expected_enabled_true_feature_off, 200),
        ("false", "ab_test1", "feature_1", expected_enabled_false_feature_on, 200),
        ("false", "ab_test1", "feature_3", expected_enabled_false_feature_off, 200),
        pytest.param("", "ab_test1", "feature_1", expected_enabled_empty, 400,
                     marks=pytest.mark.xfail(
                         reason="Define what here - status code should be 4xx")),
        pytest.param("invalid value for enabled", "ab_test1", "feature_1",
                     expected_enabled_invalid, 400,
                     marks=pytest.mark.xfail(
                         reason="Define what here - status code should be 4xx"))
    ], ids=["enabled true, all true", "enabled true, feature off",
            "enabled false, feature on",
            "enabled false, feature off", "empty value for enabled",
            "invalid value for enabled"])
def test_activate__enabled(session_obj, enabled, experimentKey, featureKey,
                           expected_response, expected_status_code):
    """
    Filter the activation response to return only enabled decisions.
    Value for enabled key needs to be a string: "true" or "false"

    - feature_1 feature is enabled - should not appear in response when enabled is set to False
    - featur_3 feature is not enabled in the project - should not appear in the project when enabled is True
    :param session_obj: session fixture
    """
    payload = '{"userId": "matjaz", "userAttributes": {"attr_1": "hola"}}'
    params = {
        "experimentKey": experimentKey,
        "featureKey": featureKey,
        "enabled": enabled
    }

    resp = session_obj.post(BASE_URL + ENDPOINT_ACTIVATE, params=params,
                            json=json.loads(payload))

    actual_response = sort_response(resp.json(), 'experimentKey', 'featureKey')
    expected_response = sort_response(json.loads(expected_response), 'experimentKey',
                                      'featureKey')
    assert actual_response == expected_response
    assert resp.status_code == expected_status_code
    resp.raise_for_status()


# #######################################################
# MISCELANEOUS ALTERNATIVE TEST CASES
# #######################################################


expected_activate_with_config = """[
  {
    "experimentKey": "ab_test1",
    "featureKey": "",
    "variationKey": "variation_1",
    "type": "experiment",
    "enabled": true
  },
  {
    "experimentKey": "feature_2_test",
    "featureKey": "",
    "variationKey": "variation_1",
    "type": "experiment",
    "enabled": true
  },
  {
    "experimentKey": "",
    "featureKey": "feature_1",
    "variationKey": "",
    "type": "feature",
    "variables": {
      "bool_var": true,
      "double_var": 5.6,
      "int_var": 1,
      "str_var": "hello"
    },
    "enabled": true
  },
  {
    "experimentKey": "",
    "featureKey": "feature_2",
    "variationKey": "",
    "type": "feature",
    "enabled": true
  },
  {
    "experimentKey": "",
    "featureKey": "feature_3",
    "variationKey": "",
    "type": "feature",
    "enabled": false
  },
  {
    "experimentKey": "",
    "featureKey": "feature_4",
    "variationKey": "",
    "type": "feature",
    "enabled": true
  },
  {
    "experimentKey": "",
    "featureKey": "feature_5",
    "variationKey": "",
    "type": "feature",
    "enabled": true
  }
]"""


def test_activate_with_config(session_obj):
    """
    Tests experimentKeys, featureKeys, variables and variations because it
    validates against the whole response body.

    In "activate"
    Request payload defines the “who” (user id and attributes)
    while the query parameters define the “what” (feature, experiment, etc)

    Request parameter is a list of experiment keys or feature keys.
    If you want both add both and separate them with comma.
    Example:
    params = {
        "featureKey": <list of feature keys>,
        "experimentKey": <list of experiment keys>
    }

    Need to sort the response (list of dictionaries). And the sorting needs to be primary
    and secondary, because we are getting response for two params - experimentKey and
    featureKey and they have different responses. experimentKey has experimentKey field
    always populated and it has featureKey empty.
    Whereas featureKey response has featureKey field populated and experimentKey empty.
    When we sort on both then the responses are properly sorted and ready for being
    asserted on.
    :param session_obj: session object
    """
    # config
    resp = session_obj.get(BASE_URL + ENDPOINT_CONFIG)
    resp_config = resp.json()

    # activate
    feat = [key for key in resp_config['featuresMap']]
    exp = [key for key in resp_config['experimentsMap']]

    payload = '{"userId": "matjaz", "userAttributes": {"attr_1": "hola"}}'
    params = {
        "featureKey": feat,
        "experimentKey": exp
    }

    resp_activate = session_obj.post(BASE_URL + ENDPOINT_ACTIVATE, params=params,
                                     json=json.loads(payload))

    resp_activate.raise_for_status()

    sorted_actual = sort_response(resp_activate.json(), 'experimentKey', 'featureKey')
    sorted_expected = sort_response(json.loads(expected_activate_with_config),
                                    'experimentKey',
                                    'featureKey')

    assert sorted_actual == sorted_expected
