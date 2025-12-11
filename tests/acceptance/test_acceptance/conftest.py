import os

import pytest
import requests

# sdk key of the project "Agent Acceptance", under QA account
sdk_key = "KZbunNn9bVfBWLpZPq2XC4"

# sdk key of the project "Agent Acceptance w ODP", under QA account
sdk_key_odp = "91GuiKYH8ZF1hLLXR7DR1"

# sdk key for holdouts datafile
sdk_key_holdouts = "BLsSFScP7tSY5SCYuKn8c"

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
def session_override_sdk_key_odp(session_obj):
    """
    Override session_obj fixture with odp SDK key.
    :param session_obj: session fixture object
    :return: updated session object
    """
    session_obj.headers['X-Optimizely-SDK-Key'] = sdk_key_odp
    return session_obj

@pytest.fixture(scope='function')
def session_override_sdk_key(session_obj):
    """
    Override session_obj fixture with invalid SDK key.
    :param session_obj: session fixture object
    :return: updated session object
    """
    session_obj.headers['X-Optimizely-SDK-Key'] = 'xxx_invalid_sdk_key_xxx'
    return session_obj


@pytest.fixture(scope='function')
def session_override_sdk_key_holdouts(session_obj):
    """
    Override session_obj fixture with holdouts SDK key.
    :param session_obj: session fixture object
    :return: updated session object
    """
    session_obj.headers['X-Optimizely-SDK-Key'] = sdk_key_holdouts
    return session_obj


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
