"""conftest.py for fixtures"""

import os
import signal
import subprocess

import pytest
import requests
from tests.acceptance.helpers import get_process_id_list
from tests.acceptance.helpers import wait_for_agent_to_start

# sdk key of the project "Agent Acceptance", under QA account
sdk_key = "KZbunNn9bVfBWLpZPq2XC4"


@pytest.fixture
def session_obj():
    """
    Using session object in each test allows to preserve headers and any other parameters
    in case we call different API endpoints in the same test.
    The Session object allows to persist certain parameters across requests (per Requests
    docs: https://requests.readthedocs.io/en/master/user/advanced/#session-objects)
    There is no harm in using this if only using one request, but it leaves it open if
    we have tests that require multiple requests.
    :return: session object
    """
    s = requests.Session()
    s.headers.update({'Content-Type': 'application/json',
                      'X-Optimizely-SDK-Key': sdk_key})
    return s


@pytest.fixture(scope='function')
def session_override_sdk_key(session_obj):
    """
    Override session_obj fixture with invalid SDK key.
    :param session_obj: session fixture object
    :return: updated session object
    """
    session_obj.headers['X-Optimizely-SDK-Key'] = 'xxx_invalid_sdk_key_xxx'
    return session_obj


@pytest.fixture(scope='session', autouse=True)
def agent_server():
    """
    Starts Agent server.
    Test run after that.
    Stops Agent server.
    Also updates config.yaml file - before agent server starts it sets enableNotifications
    to true and enableOverrides to true and then both back to false at the end
    """
    host = os.getenv('host')

    if host == 'http://localhost:8080':
        # start server
        subprocess.Popen(["make", "run"], shell=False)
        wait_for_agent_to_start()

        yield

        # Stop server
        # Kill all 'optimizely' processes
        # ('optimizely ' are processes associated with Agent server and set in ENV var?
        # here: https://github.com/optimizely/agent/blob/master/cmd/main.go#L62)
        # does not remove zombie processes though
        pid_integers = get_process_id_list('optimizely')
        for proc in pid_integers:
            os.kill(proc, signal.SIGKILL)
            print('\n========  Killing process pid', proc, end='')
    else:
        yield 


def pytest_addoption(parser):
    """
    Adding CLI option to specify host URL to run tests on.
    :param parser: parser
    """
    parser.addoption("--host", action="store", default="http://localhost:8080",
                     help="Specify host URL to run tests on.")


def pytest_configure(config):
    """
    An official pytest hook to retrieve the value of CLI option.
    https://docs.pytest.org/en/latest/reference.html#_pytest.hookspec.pytest_addoption
    See this line:
    "config.getoption(name) to retrieve the value of a command line option."
    And this ref:
    https://docs.pytest.org/en/latest/reference.html#_pytest.hookspec.pytest_configure
    Also see accepted answer here:
    https://stackoverflow.com/questions/46088297/how-do-i-access-the-command-line-input
    -in-pytest-conftest-from-the-pytest-addopt
    :param config: config
    """
    os.environ["host"] = config.getoption('host')
