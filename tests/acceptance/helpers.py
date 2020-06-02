import json
import os
import string
import time
from random import randint, choice

import requests

ENDPOINT_ACTIVATE = '/v1/activate'
ENDPOINT_CONFIG = '/v1/config'
ENDPOINT_NOTIFICATIONS = '/v1/notifications/event-stream'
ENDPOINT_OVERRIDE = '/v1/override'
ENDPOINT_TRACK = '/v1/track'

PROJECT_ROOT = os.path.dirname(os.path.abspath(__file__))


def test_health():
    """
    Checks if Agent is up.
    :return: boolean True or None
    """
    try:
        resp = requests.get('http://localhost:8080/health')
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
    Helper function to activat eexperiment.
    :param sess: API request session_object
    :return: response
    """
    BASE_URL = os.getenv('host')
    payload = {"userId": "matjaz", "userAttributes": {"attr_1": "hola"}}
    params = {"experimentKey": 'ab_test1'}
    resp = sess.post(BASE_URL + ENDPOINT_ACTIVATE, params=params, json=payload)
    return resp


def override_variation(sess, override_with):
    """
    Helper funciton to override a variation.
    :param sess: API request session object.
    :param override_with: provide new variation name as string to override with
    :return: response
    """
    BASE_URL = os.getenv('host')
    payload = {"userId": "matjaz", "userAttributes": {"attr_1": "hola"},
               "experimentKey": "ab_test1", "variationKey": f"{override_with}"}
    resp = sess.post(BASE_URL + ENDPOINT_OVERRIDE, json=payload)
    return resp
