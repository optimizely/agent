package optimizely

import (
	"encoding/json"

	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/rs/zerolog/log"

	"github.com/optimizely/agent/pkg/cluster"
	decision2 "github.com/optimizely/agent/pkg/optimizely/decision"
)

var setForcedVariationHeader = "s"
var removeForcedVariationHeader = "r"
var initConfigHeader = "i"
var updateConfigHeader = "u"

func init() {
	cluster.Listen(setForcedVariationHeader, setForcedVariation)
	cluster.Listen(removeForcedVariationHeader, removeForcedVariation)
	cluster.Listen(initConfigHeader, initConfig)
	cluster.Listen(updateConfigHeader, updateConfig)

	cluster.LocalStateFun = localState
	cluster.MergeStateFun = mergeState
}

type overrideBroadcast struct {
	SDKKey        string `json:"k"`
	UserID        string `json:"u"`
	ExperimentKey string `json:"e"`
	VariationKey  string `json:"v"`
}

type configBroadcast struct {
	SDKKey string `json:"k"`
	Type   bool   `json:"t"`
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

func parseOverride(payload []byte) (overrideBroadcast, *decision2.MapExperimentOverridesStore, error) {
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

func BroadcastInitConfig(sdkKey string) error {
	log.Info().Msgf("Broadcast init config for SDK Key: %s", sdkKey)
	request := configBroadcast{
		SDKKey: sdkKey,
	}

	return cluster.Broadcast(initConfigHeader, request)
}

func initConfig(payload []byte) {
	request := configBroadcast{}
	if err := json.Unmarshal(payload, &request); err != nil {
		log.Warn().Err(err).Msg("Unable to process config update.")
	}

	log.Info().Msgf("Received broadbast to init config for SDK Key: %s", request.SDKKey)
	if _, err := optlyCache.GetClient(request.SDKKey); err != nil {
		log.Warn().Err(err).Msg("failded to init OptlyClient via broadcast")
	}
}

func BroadcastUpdateConfig(sdkKey string) error {
	log.Info().Msgf("Broadcast update config for SDK Key: %s", sdkKey)
	request := configBroadcast{
		SDKKey: sdkKey,
	}

	return cluster.Broadcast(updateConfigHeader, request)
}

func updateConfig(payload []byte) {
	request := configBroadcast{}
	if err := json.Unmarshal(payload, &request); err != nil {
		log.Warn().Err(err).Msg("Unable to process config update.")
	}

	log.Info().Msgf("Received broadcast to update config for SDK Key: %s", request.SDKKey)
	optlyCache.UpdateConfigs(request.SDKKey)
}

type State struct {
	Configs   []configBroadcast   `json:"configs"`
	Overrides []overrideBroadcast `json:"overrides"`
}

func localState() []byte {
	state := LocalState()
	if payload, err := json.Marshal(state); err == nil {
		return payload
	} else {
		log.Warn().Err(err).Msg("failed to serialize local state.")
		return []byte{}
	}
}

func LocalState() State {
	cb := make([]configBroadcast, 0)
	ob := make([]overrideBroadcast, 0)

	for tuple := range optlyCache.optlyMap.IterBuffered() {
		sdkKey := tuple.Key

		if optlyClient, ok := tuple.Val.(*OptlyClient); !ok {
			log.Warn().Msg("not a valid OptlyClient")
		} else {
			cb = append(cb, configBroadcast{SDKKey: sdkKey})
			for k, v := range optlyClient.ForcedVariations.OverridesMap {
				override := overrideBroadcast{
					SDKKey:        sdkKey,
					UserID:        k.UserID,
					ExperimentKey: k.ExperimentKey,
					VariationKey:  v,
				}
				ob = append(ob, override)
			}
		}
	}

	return State{
		Configs:   cb,
		Overrides: ob,
	}
}

func mergeState(payload []byte) {
	log.Info().Msg("merging cluster state")
	state := &State{}
	if err := json.Unmarshal(payload, state); err != nil {
		log.Warn().Err(err).Msg("unable to parse remote state")
	}

	for _, cb := range state.Configs {
		if _, err := optlyCache.GetClient(cb.SDKKey); err != nil {
			log.Warn().Err(err).Msg("failded to init OptlyClient via broadcast")
		}
	}

	for _, ob := range state.Overrides {
		oc, err := optlyCache.GetClient(ob.SDKKey)
		if err != nil {
			log.Warn().Err(err).Msg("failded to init OptlyClient via broadcast")
			continue
		}

		forcedVariationKey := decision.ExperimentOverrideKey{
			UserID:        ob.UserID,
			ExperimentKey: ob.ExperimentKey,
		}

		oc.ForcedVariations.SetVariation(forcedVariationKey, ob.VariationKey)
	}
}
