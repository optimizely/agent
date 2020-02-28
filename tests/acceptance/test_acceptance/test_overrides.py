import os

import pytest

from tests.acceptance.helpers import ENDPOINT_OVERRIDE

BASE_URL = os.getenv('host')


# TODO - USE EXISTING EXAMPLE: https://github.com/optimizely/agent/blob/master/examples/override.py
@pytest.mark.xfail(reason='******** FIX. Says overrides not enabled ????? *********')
def test_overrides(session_obj):
    """
    Override an experiment decision for a user
    Override a experiment or feature test decision used in future user based decisions.
    This override is only stored locally in memory for debugging and testing purposes
    and should not be used for production overrides.

    Responses from the spec:
        '201':
          description: Forced variation set
        '204':
          description: Forced variation was already set
        '400':
          description: Invalid user id, experiment key, or variation key
        '403':
          $ref: '#/components/responses/Forbidden'

    :param session_obj: session fixture
    """
    payload = {"userId": "matjaz", "userAttributes": {"attr_1": "hola"},
               "experimentKey": "ab_test1", "variationKey": "new_variation"}
    resp = session_obj.post(BASE_URL + ENDPOINT_OVERRIDE, json=payload)

    assert resp.status_code == 201

# TODO - add tests - no deletion, do status code checks, param check,
#  negative status code etc
