import os

import pytest
import requests

from src.helpers import ENDPOINT_ACTIVATE
from src.helpers import ENDPOINT_CONFIG
from src.helpers import ENDPOINT_TRACK
from src.helpers import sort_response

BASE_URL = os.getenv('host')

# #######################################################
# TESTS
# #######################################################
expected_activate = [
    {'experimentKey': 'ab_test1', 'featureKey': '', 'variationKey': 'variation_1',
     'type': 'experiment', 'enabled': True},
    {'experimentKey': '', 'featureKey': 'feature_1', 'variationKey': '',
     'type': 'feature',
     'variables': {'bool_var': True, 'double_var': 5.6, 'int_var': 1, 'str_var': 'hello'},
     'enabled': True}]


# TODO - parameterize - if param is empty or invalid
def test_activate__experiment_and_feature(session_obj):
    """
    Test validates:
    1. Presence of correct variation in the returned decision for AB experiment
    2. That feature is enabled in the decision for the feature test
    Instead of on single field (variation, enabled), validation is done on the whole
    response (that includes variations and enabled fields).
    This is to add extra robustness to the test.

    Sort the reponses because dictionaries shuffle order.
    :param session_obj: session object
    """
    feature_key = 'feature_1'
    experiment_key = 'ab_test1'

    payload = {"userId": "matjaz", "userAttributes": {"attr_1": "hola"}}
    params = {
        "featureKey": feature_key,
        "experimentKey": experiment_key
    }

    resp = session_obj.post(BASE_URL + ENDPOINT_ACTIVATE, params=params, json=payload)

    resp.raise_for_status()

    sorted_actual = sort_response(resp.json(), 'experimentKey', 'featureKey')
    sorted_expected = sort_response(expected_activate, 'experimentKey', 'featureKey')

    assert sorted_actual == sorted_expected


expected_activate_type_exper = [
    {'experimentKey': 'feature_2_test', 'featureKey': '', 'variationKey': 'variation_1',
     'type': 'experiment', 'enabled': True},
    {'experimentKey': 'ab_test1', 'featureKey': '', 'variationKey': 'variation_1',
     'type': 'experiment', 'enabled': True}]
expected_activate_type_feat = [
    {'experimentKey': '', 'featureKey': 'feature_2', 'variationKey': '',
     'type': 'feature', 'enabled': True},
    {'experimentKey': '', 'featureKey': 'feature_3', 'variationKey': '',
     'type': 'feature', 'enabled': False},
    {'experimentKey': '', 'featureKey': 'feature_4', 'variationKey': '',
     'type': 'feature', 'enabled': True},
    {'experimentKey': '', 'featureKey': 'feature_5', 'variationKey': '',
     'type': 'feature', 'enabled': True},
    {'experimentKey': '', 'featureKey': 'feature_1', 'variationKey': '',
     'type': 'feature',
     'variables': {'bool_var': True, 'double_var': 5.6, 'int_var': 1, 'str_var': 'hello'},
     'enabled': True}]


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
    payload = {"userId": "matjaz", "userAttributes": {"attr_1": "hola"}}
    params = {"type": decision_type}

    resp = session_obj.post(BASE_URL + ENDPOINT_ACTIVATE, params=params, json=payload)

    if decision_type in ['experiment', 'feature']:
        sorted_actual = sort_response(resp.json(), 'experimentKey', 'featureKey')
        sorted_expected = sort_response(expected_response, 'experimentKey', 'featureKey')
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
    payload = {"userId": "matjaz", "userAttributes": {"attr_1": "hola"}}
    params = {"type": "experiment"}

    with pytest.raises(requests.exceptions.HTTPError):
        resp = session_override_sdk_key.post(BASE_URL + ENDPOINT_ACTIVATE, params=params,
                                             json=payload)

        assert resp.status_code == 403
        assert resp.json()['error'] == 'unable to fetch fresh datafile (consider ' \
                                       'rechecking SDK key), status code: 403 Forbidden'

        resp.raise_for_status()


# TODO - HOW TO DO THIS TEST ????? Check if impression event was sent in/with track?
# TODO - THIS IS "GET_VARIATION" !!!! - a version of activate that doesn't send the impression event!!!!
@pytest.mark.xfail(reason='******** To fix the test. *********')
def test_activate__disableTracking(session_obj):
    """
    Setting to true will disable impression tracking for ab experiments and feature tests.
    :param session_obj: session fixture
    """
    payload = {"userId": "matjaz", "userAttributes": {"attr_1": "hola"}}
    params = {
        "disableTracking": True
    }

    resp = session_obj.post(BASE_URL + ENDPOINT_ACTIVATE, params=params, json=payload)

    resp.raise_for_status()

    # TRACKING CHECK HERE
    params = {"eventKey": "myevent"}

    resp = session_obj.post(BASE_URL + ENDPOINT_TRACK, params=params,
                            json=payload)  # TODO - THIS SHOULD NOT SEND AN EVENT when disableTracking is True. BUT IT SENDS !!!!!!!!!!!!!!!
    assert not resp.status_code == 204


# TODO - I DON'T GET THIS, or it's a bug
# TODO: NOT WORKING - I GET BACK BOTH DECISIONS FOR FEATURES AND EXPERIMENTS, REGARDLESS IF THEY ARE ENABLED OR DISABLED!!!
# TODO - SHOULD FILTER PER "enabled" field.
def test_activate__enabled(session_obj):
    """
    Filter the activation response to return only enabled decisions.
    :param session_obj: session fixture
    """
    feature_1_key = 'feature_1'  # should not appear in response when enabled is set to False
    feature_3_key = 'feature_3'  # TODO - THIS FEATURE 3 IS NOT ENABLED IN THE PROJECT - SHOULD NOT APPEAR IN THE RESPONSE WHEN enabled is True
    experiment_key = 'ab_test1'

    payload = {"userId": "matjaz", "userAttributes": {"attr_1": "hola"}}
    params = {
        "experimentKey": experiment_key,
        "featureKey": feature_3_key,
        "enabled": True
    }

    resp = session_obj.post(BASE_URL + ENDPOINT_ACTIVATE, params=params, json=payload)

    resp.raise_for_status()


# #######################################################
# MISCELANEOUS ALTERNATIVE TEST CASES
# #######################################################


expected_activate_with_config = [
    {'experimentKey': 'ab_test1', 'featureKey': '', 'variationKey': 'variation_1',
     'type': 'experiment', 'enabled': True},
    {'experimentKey': 'feature_2_test', 'featureKey': '', 'variationKey': 'variation_1',
     'type': 'experiment', 'enabled': True},
    {'experimentKey': '', 'featureKey': 'feature_1', 'variationKey': '',
     'type': 'feature',
     'variables': {'bool_var': True, 'double_var': 5.6, 'int_var': 1, 'str_var': 'hello'},
     'enabled': True},
    {'experimentKey': '', 'featureKey': 'feature_2', 'variationKey': '',
     'type': 'feature', 'enabled': True},
    {'experimentKey': '', 'featureKey': 'feature_3', 'variationKey': '',
     'type': 'feature', 'enabled': False},
    {'experimentKey': '', 'featureKey': 'feature_4', 'variationKey': '',
     'type': 'feature', 'enabled': True},
    {'experimentKey': '', 'featureKey': 'feature_5', 'variationKey': '',
     'type': 'feature', 'enabled': True}]


def test_activate_with_config(session_obj):
    """
    Tests experimentKeys, featureKeys, variables and variations because it
    valdiates against the whole response body.

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
    # TODO make a helper function here for "get config" - will it be used a lot?
    resp = session_obj.get(BASE_URL + ENDPOINT_CONFIG)
    resp_config = resp.json()

    # activate
    feat = [key for key in resp_config['featuresMap']]
    exp = [key for key in resp_config['experimentsMap']]

    payload = {"userId": "matjaz", "userAttributes": {"attr_1": "hola"}}
    params = {
        "featureKey": feat,
        "experimentKey": exp
    }

    resp_activate = session_obj.post(BASE_URL + ENDPOINT_ACTIVATE, params=params,
                                     json=payload)

    resp_activate.raise_for_status()

    sorted_actual = sort_response(resp_activate.json(), 'experimentKey', 'featureKey')
    sorted_expected = sort_response(expected_activate_with_config, 'experimentKey',
                                    'featureKey')

    assert sorted_actual == sorted_expected
