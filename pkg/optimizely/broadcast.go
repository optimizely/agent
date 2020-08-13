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

	log.Info().Msgf("Broadcast SET override for userID: %s", request.UserID)
	return cluster.Broadcast(setForcedVariationHeader, request)
}

func setForcedVariation(payload []byte) {
	request, overrides, err := parseOverride(payload)
	if err != nil {
		log.Warn().Err(err).Msg("Error parsing setForcedVariation request")
	}

	log.Info().Msgf("Received broadcast to SET override for userID: %s", request.UserID)
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

	log.Info().Msgf("Broadcast REMOVE override for userID: %s", request.UserID)
	return cluster.Broadcast(setForcedVariationHeader, request)
}

func removeForcedVariation(payload []byte) {
	request, overrides, err := parseOverride(payload)
	if err != nil {
		log.Warn().Err(err).Msg("Error parsing setForcedVariation request")
	}

	log.Info().Msgf("Received broadcast to REMOVE override for userID: %s", request.UserID)
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
	log.Info().Msgf("Received broadbast to update config for SDK Key: %s", sdkKey)
	optlyCache.UpdateConfigs(sdkKey)
}
