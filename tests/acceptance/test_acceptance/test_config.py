import json

import pytest
import requests

from tests.acceptance.helpers import ENDPOINT_CONFIG
from tests.acceptance.helpers import create_and_validate_request_and_response

expected_config = """{
  "revision": "111",
  "attributes": [
    {
      "id": "16921322086",
      "key": "attr_1"
    }
  ],
  "audiences": [
    {
      "conditions": "[\"and \", [\"or \", [\"or \", {\"match \": \"exact \", \"name \": \"attr_1\",\"type\": \"custom_attribute\",\"value\": \"hola\"}]]]",
      "id": "16902921321",
      "name": "Audience1"
    }
  ],
  "environmentKey": "",
  "events": [
    {
      "experimentIds": [
        "16910084756",
        "16911963060"
      ],
      "id": "16911532385",
      "key": "myevent"
    }
  ],
  "experimentsMap": {
    "ab_test1": {
      "audiences": "\"Audience1\"",
      "id": "16911963060",
      "key": "ab_test1",
      "variationsMap": {
        "variation_1": {
          "id": "16905941566",
          "key": "variation_1",
          "featureEnabled": false,
          "variablesMap": {
            
          }
        },
        "variation_2": {
          "id": "16927770169",
          "key": "variation_2",
          "featureEnabled": false,
          "variablesMap": {
            
          }
        }
      }
    },
    "feature_2_test": {
      "audiences": "\"Audience1\"",
      "id": "16910084756",
      "key": "feature_2_test",
      "variationsMap": {
        "variation_1": {
          "id": "16925360560",
          "key": "variation_1",
          "featureEnabled": true,
          "variablesMap": {
            
          }
        },
        "variation_2": {
          "id": "16915611472",
          "key": "variation_2",
          "featureEnabled": true,
          "variablesMap": {
            
          }
        }
      }
    }
  },
  "featuresMap": {
    "feature_1": {
      "deliveryRules": [
        {
          "audiences": "\"Audience1\"",
          "id": "16941022436",
          "key": "16941022436",
          "variationsMap": {
            "16906801184": {
              "featureEnabled": true,
              "id": "16906801184",
              "key": "16906801184",
              "variablesMap": {
                "bool_var": {
                  "id": "16932993089",
                  "key": "bool_var",
                  "type": "boolean",
                  "value": "true"
                },
                "double_var": {
                  "id": "16923002469",
                  "key": "double_var",
                  "type": "double",
                  "value": "5.6"
                },
                "int_var": {
                  "id": "16937161477",
                  "key": "int_var",
                  "type": "integer",
                  "value": "1"
                },
                "str_var": {
                  "id": "16916052157",
                  "key": "str_var",
                  "type": "string",
                  "value": "hello"
                }
              }
            }
          }
        }
      ],
      "experimentRules": [
        
      ],
      "id": "16925981047",
      "key": "feature_1",
      "experimentsMap": {
        
      },
      "variablesMap": {
        "bool_var": {
          "id": "16932993089",
          "key": "bool_var",
          "type": "boolean",
          "value": "true"
        },
        "double_var": {
          "id": "16923002469",
          "key": "double_var",
          "type": "double",
          "value": "5.6"
        },
        "int_var": {
          "id": "16937161477",
          "key": "int_var",
          "type": "integer",
          "value": "1"
        },
        "str_var": {
          "id": "16916052157",
          "key": "str_var",
          "type": "string",
          "value": "hello"
        }
      }
    },
    "feature_2": {
      "id": "16928980973",
      "key": "feature_2",
      "deliveryRules": [
        {
          "audiences": "\"Audience1\"",
          "id": "16924931120",
          "key": "16924931120",
          "variationsMap": {
            "16931381940": {
              "featureEnabled": true,
              "id": "16931381940",
              "key": "16931381940",
              "variablesMap": {
                
              }
            }
          }
        }
      ],
      "experimentRules": [
        {
          "audiences": "\"Audience1\"",
          "id": "16910084756",
          "key": "feature_2_test",
          "variationsMap": {
            "variation_1": {
              "featureEnabled": true,
              "id": "16925360560",
              "key": "variation_1",
              "variablesMap": {
                
              }
            },
            "variation_2": {
              "featureEnabled": true,
              "id": "16915611472",
              "key": "variation_2",
              "variablesMap": {
                
              }
            }
          }
        }
      ],
      "experimentsMap": {
        "feature_2_test": {
          "audiences": "\"Audience1\"",
          "id": "16910084756",
          "key": "feature_2_test",
          "variationsMap": {
            "variation_1": {
              "id": "16925360560",
              "key": "variation_1",
              "featureEnabled": true,
              "variablesMap": {
                
              }
            },
            "variation_2": {
              "id": "16915611472",
              "key": "variation_2",
              "featureEnabled": true,
              "variablesMap": {
                
              }
            }
          }
        }
      },
      "variablesMap": {
        
      }
    },
    "feature_3": {
      "deliveryRules": [
        {
          "audiences": "",
          "id": "16907440927",
          "key": "16907440927",
          "variationsMap": {
            "16908510336": {
              "featureEnabled": false,
              "id": "16908510336",
              "key": "16908510336",
              "variablesMap": {
                
              }
            }
          }
        }
      ],
      "experimentRules": [
        
      ],
      "id": "16907463855",
      "key": "feature_3",
      "experimentsMap": {
        
      },
      "variablesMap": {
        
      }
    },
    "feature_4": {
      "deliveryRules": [
        {
          "audiences": "",
          "id": "16939051724",
          "key": "16939051724",
          "variationsMap": {
            "16925940659": {
              "featureEnabled": true,
              "id": "16925940659",
              "key": "16925940659",
              "variablesMap": {
                
              }
            }
          }
        }
      ],
      "experimentRules": [
        
      ],
      "id": "16912161768",
      "key": "feature_4",
      "experimentsMap": {
        
      },
      "variablesMap": {
        
      }
    },
    "feature_5": {
      "deliveryRules": [
        {
          "audiences": "",
          "id": "16932940705",
          "key": "16932940705",
          "variationsMap": {
            "16927890136": {
              "featureEnabled": true,
              "id": "16927890136",
              "key": "16927890136",
              "variablesMap": {
                
              }
            }
          }
        }
      ],
      "experimentRules": [
        
      ],
      "id": "16923312421",
      "key": "feature_5",
      "experimentsMap": {
        
      },
      "variablesMap": {
        
      }
    }
  },
  "sdkKey": ""
}"""


def test_config(session_obj):
    """
    Test validates all returned available experiment and features definitions
    for this environment.

    Note: Test will fail as soon as anything in the response body is modified.
    If someone updates any of the fields, the expected_response will need to be updated
    as well.
    :param session_obj: session object
    """
    resp = create_and_validate_request_and_response(ENDPOINT_CONFIG, 'get', session_obj)
    assert resp.status_code == 200
    resp.raise_for_status()
    assert json.loads(expected_config) == resp.json()


def test_config_403(session_override_sdk_key):
    """
    Test that 403 Forbidden is returned. We use invalid SDK key to trigger 403.
    :param : session_obj
    """
    with pytest.raises(requests.exceptions.HTTPError):
        resp = create_and_validate_request_and_response(ENDPOINT_CONFIG, 'get', session_override_sdk_key)

        assert resp.status_code == 403
        assert resp.json()['error'] == 'unable to fetch fresh datafile (consider ' \
                                       'rechecking SDK key), status code: 403 Forbidden'

        resp.raise_for_status()
