import json
import os
import string
from random import randint, choice

import yaml
from openapi_core import create_spec
from openapi_core.validation.request.datatypes import (OpenAPIRequest, RequestParameters)
from openapi_core.validation.request.validators import RequestValidator
from openapi_core.validation.response.datatypes import OpenAPIResponse
from openapi_core.validation.response.validators import ResponseValidator
from werkzeug.datastructures import ImmutableMultiDict

ENDPOINT_ACTIVATE = '/v1/activate'
ENDPOINT_CONFIG = '/v1/config'
ENDPOINT_NOTIFICATIONS = '/v1/notifications/event-stream'
ENDPOINT_OVERRIDE = '/v1/override'
ENDPOINT_TRACK = '/v1/track'
ENDPOINT_BATCH = '/v1/batch'
ENDPOINT_DECIDE = '/v1/decide'
ENDPOINT_DATAFILE = '/v1/datafile'
ENDPOINT_SAVE = '/v1/save'
ENDPOINT_LOOKUP = '/v1/lookup'
ENDPOINT_SEND_ODP_EVENT = '/v1/send-odp-event'

YAML_FILE_PATH = os.getenv('OPENAPI_YAML_PATH', 'api/openapi-spec/openapi.yaml')


def parse_yaml(path):
    with open(path, 'r') as stream:
        try:
            return yaml.safe_load(stream)
        except yaml.YAMLError as exc:
            print(exc)


spec = create_spec(parse_yaml(YAML_FILE_PATH))


def get_random_string():
    """
    :return: randomized string
    """
    return "".join(choice(string.ascii_letters) for _ in range(randint(10, 15)))


def get_pretty_json(dictionary, spaces=4):
    """
    Makes JSON output prettier and readable.
    :return: stringified JSON
    """
    return json.dumps(dictionary, indent=spaces)


def sort_response(response_dict, *args):
    """
    Used in tests to sort responses by two or more keys.
    For example if response includes experimentKey and FeatureKey, the function
    will sort by primary and secondary key, depending which one you put first.
    The first param will be primary sorted, second secondary.
    Can handle arbitrary number of arguments.
    :param response_dict: response
    :param args: usually experimentKey and featureKey
    :return: sorted response
    """
    return sorted(response_dict, key=lambda k: tuple(map(k.__getitem__, args)))


# Helper funcitons for overrides
def activate_experiment(sess):
    """
    Helper function to activate experiment.
    :param sess: API request session_object
    :return: response
    """
    payload = '{"userId": "matjaz", "userAttributes": {"attr_1": "hola"}}'
    params = {"experimentKey": "ab_test1"}

    resp = create_and_validate_request_and_response(ENDPOINT_ACTIVATE, 'post', sess, payload=payload, params=params)

    return resp


def override_variation(sess, override_with):
    """
    Helper funciton to override a variation.
    :param sess: API request session object.
    :param override_with: provide new variation name as string to override with
    :return: response
    """
    payload = {"userId": "matjaz", "userAttributes": {"attr_1": "hola"},
               "experimentKey": "ab_test1", "variationKey": f"{override_with}"}

    resp = create_and_validate_request_and_response(
        ENDPOINT_OVERRIDE, 'post', sess, payload=json.dumps(payload)
    )

    return resp


def create_and_validate_request(endpoint, method, payload='', params=[], headers=[]):
    """
    Helper function to create OpenAPIRequest and validate it
    :param endpoint: API endpoint
    :param method: API request method
    :param payload: API request payload
    :param params: API request payload
    :param headers: API request headers
    :return:
        - request: OpenAPIRequest
        - request_result: result of request validation
    """
    parameters = RequestParameters(
        query=ImmutableMultiDict(params),
        path=endpoint,
        header=headers
    )

    request = OpenAPIRequest(
        full_url_pattern=endpoint,
        method=method,
        parameters=parameters,
        body=payload,
        mimetype='application/json',
    )

    validator = RequestValidator(spec)
    request_result = validator.validate(request)

    return request, request_result


def create_and_validate_response(request, response):
    """
    Helper function to create OpenAPIResponse and validate it
    :param request: OpenAPIRequest
    :param response: API response
    :return:
        - result: result of response validation
    """
    response = OpenAPIResponse(
        data=response.content,
        status_code=response.status_code,
        mimetype='application/json'
    )

    validator = ResponseValidator(spec)
    result = validator.validate(request, response)
    return result


def create_and_validate_request_and_response(endpoint, method, session, bypass_validation_request=False,
                                             bypass_validation_response=False, payload='', params=[]):
    """
    Helper function to create OpenAPIRequest, OpenAPIResponse and validate both
    :param endpoint: API endpoint
    :param session: API valid session object
    :param bypass_validation_request: Flag to bypass request validation of invalid requests
    :param bypass_validation_response: Flag to bypass request validation of invalid responses
    :param method: API request method
    :param payload: API request payload
    :param params: API request payload
    :return:
        - response: API response object
    """
    request, request_result = create_and_validate_request(
        endpoint, method, payload, params, dict(session.headers)
    )

    if not bypass_validation_request:
        request_result.raise_for_errors()

    base_url = os.getenv('host')

    if method == 'post':
        response = session.post(base_url + endpoint, params=params, data=payload)
    elif method == 'get':
        response = session.get(base_url + endpoint, params=params, data=payload)
    response_result = create_and_validate_response(request, response)

    if not bypass_validation_response:
        response_result.raise_for_errors()

    return response
