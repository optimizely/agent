from tests.acceptance.helpers import ENDPOINT_DATAFILE
from tests.acceptance.helpers import create_and_validate_request_and_response


expected_response = {"accountId": "10845721364", "anonymizeIP": True,
                       "attributes": [{"id": "16921322086", "key": "attr_1"}], "audiences": [
        {
            "conditions": "[\"and\", [\"or\", [\"or\", {\"match\": \"exact\", \"name\": \"attr_1\", \"type\": \"custom_attribute\", \"value\": \"hola\"}]]]",
            "id": "16902921321", "name": "Audience1"}, {
            "conditions": "[\"or\", {\"match\": \"exact\", \"name\": \"$opt_dummy_attribute\", \"type\": \"custom_attribute\", \"value\": \"$opt_dummy_value\"}]",
            "id": "$opt_dummy_audience", "name": "Optimizely-Generated Audience for Backwards Compatibility"}],
                       "botFiltering": False,
                       "events": [
                           {"experimentIds": ["16910084756", "16911963060"], "id": "16911532385", "key": "myevent"}],
                       "experiments": [
                           {"audienceConditions": ["or", "16902921321"], "audienceIds": ["16902921321"],
                            "forcedVariations": {},
                            "id": "16910084756", "key": "feature_2_test", "layerId": "16933431472", "status": "Running",
                            "trafficAllocation": [{"endOfRange": 5000, "entityId": "16925360560"},
                                                  {"endOfRange": 10000, "entityId": "16925360560"}],
                            "variations": [
                                {"featureEnabled": True, "id": "16925360560", "key": "variation_1", "variables": []},
                                {"featureEnabled": True, "id": "16915611472", "key": "variation_2", "variables": []}]},
                           {"audienceConditions": ["or", "16902921321"], "audienceIds": ["16902921321"],
                            "forcedVariations": {},
                            "id": "16911963060", "key": "ab_test1", "layerId": "16916031507", "status": "Running",
                            "trafficAllocation": [{"endOfRange": 1000, "entityId": "16905941566"},
                                                  {"endOfRange": 5000, "entityId": "16905941566"},
                                                  {"endOfRange": 8000, "entityId": "16905941566"},
                                                  {"endOfRange": 9000, "entityId": "16905941566"},
                                                  {"endOfRange": 10000, "entityId": "16905941566"}],
                            "variations": [{"id": "16905941566", "key": "variation_1", "variables": []},
                                           {"id": "16927770169", "key": "variation_2", "variables": []}]}],
                       "featureFlags": [
                           {"experimentIds": [], "id": "16907463855", "key": "feature_3", "rolloutId": "16909553406",
                            "variables": []},
                           {"experimentIds": [], "id": "16912161768", "key": "feature_4", "rolloutId": "16943340293",
                            "variables": []},
                           {"experimentIds": [], "id": "16923312421", "key": "feature_5", "rolloutId": "16917103311",
                            "variables": []},
                           {"experimentIds": [], "id": "16925981047", "key": "feature_1", "rolloutId": "16928980969",
                            "variables": [
                                {"defaultValue": "hello", "id": "16916052157", "key": "str_var", "type": "string"},
                                {"defaultValue": "5.6", "id": "16923002469", "key": "double_var", "type": "double"},
                                {"defaultValue": "true", "id": "16932993089", "key": "bool_var", "type": "boolean"},
                                {"defaultValue": "1", "id": "16937161477", "key": "int_var", "type": "integer"}]},
                           {"experimentIds": ["16910084756"], "id": "16928980973", "key": "feature_2",
                            "rolloutId": "16917900798",
                            "variables": []}], "groups": [], "projectId": "16931203314", "revision": "111",
                       "rollouts": [{"experiments": [
                           {"audienceIds": [], "forcedVariations": {}, "id": "16907440927", "key": "16907440927",
                            "layerId": "16909553406",
                            "status": "Not started",
                            "trafficAllocation": [{"endOfRange": 0, "entityId": "16908510336"}],
                            "variations": [{"featureEnabled": False, "id": "16908510336", "key": "16908510336",
                                            "variables": []}]}],
                           "id": "16909553406"},
                           {"experiments": [{
                               "audienceIds": [],
                               "forcedVariations": {},
                               "id": "16932940705",
                               "key": "16932940705",
                               "layerId": "16917103311",
                               "status": "Running",
                               "trafficAllocation": [
                                   {
                                       "endOfRange": 10000,
                                       "entityId": "16927890136"}],
                               "variations": [
                                   {
                                       "featureEnabled": True,
                                       "id": "16927890136",
                                       "key": "16927890136",
                                       "variables": []}]}],
                               "id": "16917103311"},
                           {"experiments": [{
                               "audienceConditions": [
                                   "or",
                                   "16902921321"],
                               "audienceIds": [
                                   "16902921321"],
                               "forcedVariations": {},
                               "id": "16924931120",
                               "key": "16924931120",
                               "layerId": "16917900798",
                               "status": "Running",
                               "trafficAllocation": [
                                   {
                                       "endOfRange": 10000,
                                       "entityId": "16931381940"}],
                               "variations": [
                                   {
                                       "featureEnabled": True,
                                       "id": "16931381940",
                                       "key": "16931381940",
                                       "variables": []}]}],
                               "id": "16917900798"},
                           {"experiments": [{
                               "audienceConditions": [
                                   "or",
                                   "16902921321"],
                               "audienceIds": [
                                   "16902921321"],
                               "forcedVariations": {},
                               "id": "16941022436",
                               "key": "16941022436",
                               "layerId": "16928980969",
                               "status": "Running",
                               "trafficAllocation": [
                                   {
                                       "endOfRange": 10000,
                                       "entityId": "16906801184"}],
                               "variations": [
                                   {
                                       "featureEnabled": True,
                                       "id": "16906801184",
                                       "key": "16906801184",
                                       "variables": [
                                           {
                                               "id": "16932993089",
                                               "value": "true"},
                                           {
                                               "id": "16923002469",
                                               "value": "5.6"},
                                           {
                                               "id": "16916052157",
                                               "value": "hello"},
                                           {
                                               "id": "16937161477",
                                               "value": "1"}]}]}],
                               "id": "16928980969"},
                           {"experiments": [{
                               "audienceIds": [],
                               "forcedVariations": {},
                               "id": "16939051724",
                               "key": "16939051724",
                               "layerId": "16943340293",
                               "status": "Running",
                               "trafficAllocation": [
                                   {
                                       "endOfRange": 10000,
                                       "entityId": "16925940659"}],
                               "variations": [
                                   {
                                       "featureEnabled": True,
                                       "id": "16925940659",
                                       "key": "16925940659",
                                       "variables": []}]}],
                               "id": "16943340293"}],
                       "typedAudiences": [], "variables": [], "version": "4"}


def test_datafile_200(session_obj):
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
