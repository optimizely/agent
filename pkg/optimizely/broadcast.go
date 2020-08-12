package optimizely

import (
	"encoding/json"

	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/rs/zerolog/log"

	"github.com/optimizely/agent/pkg/cluster"
)

var setForcedVariationHeader = "s"
var removeForcedVariationHeader = "r"
var updateConfigHeader = "u"

func init() {
	cluster.Listen(setForcedVariationHeader, setForcedVariation)
	cluster.Listen(removeForcedVariationHeader, removeForcedVariation)
	cluster.Listen(updateConfigHeader, updateConfig)
}

type overrideBroadcast struct {
	SDKKey        string `json:"k"`
	UserID        string `json:"u"`
	ExperimentKey string `json:"e"`
	VariationKey  string `json:"v"`
}

func BroadcastSetForcedVariation(sdkKey, userID, experimentKey, variationKey string) error {
	request := overrideBroadcast{
		SDKKey:        sdkKey,
		UserID:        userID,
		ExperimentKey: experimentKey,
		VariationKey:  variationKey,
	}

	return cluster.Broadcast(setForcedVariationHeader, request)
}

func setForcedVariation(payload []byte) {
	request, overrides, err := parseOverride(payload)
	if err != nil {
		log.Warn().Err(err).Msg("Error parsing setForcedVariation request")
	}

	forcedVariationKey := decision.ExperimentOverrideKey{
		UserID:        request.UserID,
		ExperimentKey: request.ExperimentKey,
	}

	overrides.SetVariation(forcedVariationKey, request.VariationKey)
}

func BroadcastRemoveForcedVariation(sdkKey, userID, experimentKey string) error {
	request := overrideBroadcast{
		SDKKey:        sdkKey,
		UserID:        userID,
		ExperimentKey: experimentKey,
	}

	return cluster.Broadcast(setForcedVariationHeader, request)
}

func removeForcedVariation(payload []byte) {
	request, overrides, err := parseOverride(payload)
	if err != nil {
		log.Warn().Err(err).Msg("Error parsing setForcedVariation request")
	}

	forcedVariationKey := decision.ExperimentOverrideKey{
		UserID:        request.UserID,
		ExperimentKey: request.ExperimentKey,
	}

	overrides.RemoveVariation(forcedVariationKey)
}

func parseOverride(payload []byte) (overrideBroadcast, *decision.MapExperimentOverridesStore, error) {
	request := overrideBroadcast{}
	if err := json.Unmarshal(payload, &request); err != nil {
		return request, nil, err
	}

	client, err := optlyCache.GetClient(request.SDKKey)
	if err != nil {
		return request, nil, err
	}

	return request, client.ForcedVariations, nil
}

func BroadcastUpdateConfig(sdkKey string) error {
	return cluster.Broadcast(updateConfigHeader, sdkKey)
}

func updateConfig(payload []byte) {
	sdkKey := string(payload)
	optlyCache.UpdateConfigs(sdkKey)
}
