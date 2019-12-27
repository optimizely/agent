package config

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	viper.SetConfigFile("./testdata/default.yaml")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	assert.NoError(t, err)

	conf := AgentConfig{}
	err = viper.Unmarshal(&conf)
	assert.NoError(t, err)

	assert.Equal(t, 5*time.Second, conf.Server.ReadTimeout)
	assert.Equal(t, 10*time.Second, conf.Server.WriteTimeout)

	assert.True(t, conf.Log.Pretty)
	assert.Equal(t, "debug", conf.Log.Level)

	assert.False(t, conf.Admin.Enabled)
	assert.Equal(t, "3002", conf.Admin.Port)

	assert.True(t, conf.Api.Enabled)
	assert.Equal(t, 100, conf.Api.MaxConns)
	assert.Equal(t, "3000", conf.Api.Port)

	assert.True(t, conf.Webhook.Enabled)
	assert.Equal(t, "3001", conf.Webhook.Port)
	assert.Equal(t, "secret-10000", conf.Webhook.Projects[10000].Secret)
	assert.Equal(t, []string{"aaa", "bbb", "ccc"}, conf.Webhook.Projects[10000].SDKKeys)
	assert.True(t, conf.Webhook.Projects[10000].SkipSignatureCheck)
	assert.Equal(t, "secret-20000", conf.Webhook.Projects[20000].Secret)
	assert.Equal(t, []string{"xxx", "yyy", "zzz"}, conf.Webhook.Projects[20000].SDKKeys)
	assert.False(t, conf.Webhook.Projects[20000].SkipSignatureCheck)

	viper.Reset()
}
