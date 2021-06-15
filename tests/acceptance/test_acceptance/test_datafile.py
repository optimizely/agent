from tests.acceptance.helpers import ENDPOINT_DATAFILE, create_and_validate_request_and_response
from tests.acceptance.datafile import datafile as expected_response


def test_datafile_success(session_obj):
    """
    Normally a good practice is to have expected response as a string like in other tests.
    Here we are exceptionally making expected response a dict for easier comparison.
    String was causing some issues with extra white space characters.
    :param session_obj: session object
    """
    payload = '{"userId": "matjaz", "userAttributes": {"attr_1": "hola"}}'
    params = {"featureKey": "feature_1"}

    resp = create_and_validate_request_and_response(ENDPOINT_DATAFILE, 'get', session_obj,
                                                    bypass_validation_request=False,
                                                    payload=payload, params=params)

    assert expected_response == resp.json()
    assert resp.status_code == 200, resp.text
