package redis

import (
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/rs/zerolog/log"
)

// ExperimentOverrideStore implements a decision.ExperimentOverrideStore backed by redis
type ExperimentOverrideStore struct {
	client     *redis.Client
	expiration time.Duration
}

// GetVariation returns the override variation key associated with the given user+experiment key
func (s *ExperimentOverrideStore) GetVariation(overrideKey decision.ExperimentOverrideKey) (string, bool) {
	redisKey := getOverrideKeyAsString(overrideKey)
	variation, err := s.client.Get(redisKey).Result()
	if err != nil {
		log.Error().Err(err).Str("redisKey", redisKey).Msg("Unable to fetch override value")
		return "", false
	}

	log.Debug().Str("variation", variation).Str("redisKey", redisKey).Msg("get experiment override")
	return variation, variation != ""
}

func (s *ExperimentOverrideStore) SetVariation(overrideKey decision.ExperimentOverrideKey, variation string) error {
	redisKey := getOverrideKeyAsString(overrideKey)
	set, err := s.client.Set(redisKey, variation, s.expiration).Result()
	if err != nil {
		log.Error().Err(err).Str("redisKey", redisKey).Msg("Unable to fetch override value")
		return err
	}

	log.Debug().Str("set", set).Str("redisKey", redisKey).Msg("set experiment override")

	return nil
}

func getOverrideKeyAsString(overrideKey decision.ExperimentOverrideKey) string {
	return overrideKey.UserID + "|" + overrideKey.ExperimentKey
}
