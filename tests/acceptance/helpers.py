import json
import os
import string
import time
from random import randint, choice
import requests
import yaml

from openapi_core import create_spec
from openapi_core.validation.request.validators import RequestValidator
from openapi_core.validation.response.validators import ResponseValidator
from openapi_core.validation.request.datatypes import (OpenAPIRequest, RequestParameters)
from openapi_core.validation.response.datatypes import OpenAPIResponse
from werkzeug.datastructures import ImmutableMultiDict


ENDPOINT_ACTIVATE = '/v1/activate'
ENDPOINT_CONFIG = '/v1/config'
ENDPOINT_NOTIFICATIONS = '/v1/notifications/event-stream'
ENDPOINT_OVERRIDE = '/v1/override'
ENDPOINT_TRACK = '/v1/track'

YAML_FILE_PATH = os.getenv('OPENAPI_YAML_PATH', 'api/openapi-spec/openapi.yaml')

spec_dict = None
with open(YAML_FILE_PATH, 'r') as stream:
    try:
        spec_dict = yaml.safe_load(stream)
    except yaml.YAMLError as exc:
        print(exc)

spec = create_spec(spec_dict)


def test_health():
    """
    Checks if Agent is up.
    :return: boolean True or None
    """
    try:
        resp = requests.get('http://localhost:8088/health')
        if resp.json()['status'] == 'ok':
            return True
    except requests.exceptions.ConnectionError:
        print(f'Agent server is not yet ready (connection refused).')


def get_pid(name):
    """
    Gets PID of a running process by name
    :return: pid of selected process by name
    """
    from subprocess import check_output

    pid = check_output(["pidof", name])

    if not pid:
        raise ValueError(f"Pid of {name} not found. "
                         f"Likely {name} background process is not running.",
                         f"pid: {pid}")

    return pid


def wait_for_agent_to_start():
    """
    Waits until agent server is up - meaning if the port 8080 is open.
    Keeps checking in a loop if port is available
    """
    timeout = time.time() + 30  # 30 s timeout limit to prevent infinite loop

    while not test_health():
        if time.time() > timeout:
            raise RuntimeError("Timeout exceeded. Agent server not started?")
        else:
            time.sleep(1)

    print('Agent server is up and ready on localhost.')


def wait_for_agent_to_stop():
    """
    Waits until agent server stopped.
    Keeps checking in a loop if port is available
    """
    timeout = time.time() + 30  # 30 s timeout limit to prevent infinite loop

    while test_health():
        if time.time() > timeout:
            raise RuntimeError("Timeout exceeded. Agent server not started?")
        else:
            time.sleep(1)

    print('Agent server has stopped.')


def get_process_id_list(name):
    """
    Converts bytes of process id's returned from get_pid() function
    and returns a list of PIDs as integers
    :param name: name of process
    :return: list of PIDs as integers
    """
    server_processes_in_bytes = get_pid(name).decode("utf-8").strip()
    pid_list = server_processes_in_bytes.split(',')
    split_list = pid_list[0].split()
    pid_integers = [int(x) for x in split_list]
    return pid_integers


def get_random_string():
    """
    :return: randomized string
    """
    return "".join(choice(string.ascii_letters) for _ in range(randint(10, 15)))


def get_pretty_json(dictionary, spaces=4):
    """
    Makes JSON output prettuer and readable.
    :return: stringified JSON
    """
    return json.dumps(dictionary, indent=spaces)


def sort_response(response_dict, *args):
    """
    Used in tests to sort responses byt tw or more fields.
    For example if rersponse includes experimentKey and FeatureKey, the function
    will sort by primary and secondary key, depending which one you put first.
    The first param will be primary sorted, second secondary.
    :param response_dict: response
    :param args: usually experimentKey and featureKey
    :return: sorted response
    """
    return sorted(response_dict, key=lambda k: (k[args[0]], k[args[1]]))


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


def create_and_validate_request(endpoint, method, payload='', params=[]):
    """
    Helper function to create OpenAPIRequest and validate it
    :param endpoint: API endpoint
    :param method: API request method
    :param payload: API request payload
    :param params: API request payload
    :return:
        - request: OpenAPIRequest
        - request_result: result of request validation
    """
    parameters = RequestParameters(
        query=ImmutableMultiDict(params),
        path=endpoint
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


def create_and_validate_request_and_response(endpoint, method, session, bypass_validation=False, payload='', params=[]):
    """
    Helper function to create OpenAPIRequest, OpenAPIResponse and validate both
    :param endpoint: API endpoint
    :param session: API valid session object
    :param bypass_validation: Flag to bypass request validation of invalid requests
    :param method: API request method
    :param payload: API request payload
    :param params: API request payload
    :return:
        - response: API response object
    """
    request, request_result = create_and_validate_request(endpoint, method, payload, params)
    if not bypass_validation:
        pass
        # raise errors if request invalid
        request_result.raise_for_errors()

    BASE_URL = os.getenv('host')

    if method == 'post':
        response = session.post(BASE_URL + endpoint, params=params, data=payload)
    elif method == 'get':
        response = session.get(BASE_URL + endpoint, params=params, data=payload)

    response_result = create_and_validate_response(request, response)
    # raise errors if response invalid
    response_result.raise_for_errors()

    return response
