package redis

import (
	"testing"
	"time"

	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/stretchr/testify/assert"

	"github.com/optimizely/sidedoor/config"
)

func TestSetAndGet(t *testing.T) {
	store := getStore()
	key := decision.ExperimentOverrideKey{
		ExperimentKey: "test",
		UserID:        "user",
	}

	expected := "var"
	err := store.SetVariation(key, expected)
	assert.NoError(t, err)

	actual, ok := store.GetVariation(key)
	assert.True(t, ok)
	assert.Equal(t, expected, actual)

}

func getStore() *ExperimentOverrideStore {
	conf := config.RedisConfig{
		Address:  "localhost:6379",
		Password: "", // no password set
		Database: 0,  // use default DB
	}

	return &ExperimentOverrideStore{
		client:     NewClient(conf),
		expiration: 0 * time.Second,
	}
}
