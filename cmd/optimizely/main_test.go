/****************************************************************************
 * Copyright 2019-2020,2022-2023, Optimizely, Inc. and contributors         *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    http://www.apache.org/licenses/LICENSE-2.0                            *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

package main

import (
	"os"
	"testing"
	"time"

	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/optimizely"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func assertRoot(t *testing.T, actual *config.AgentConfig) {
	assert.Equal(t, "0.1.0", actual.Version)
	assert.Equal(t, "Optimizely Inc.", actual.Author)
	assert.Equal(t, "optimizely", actual.Name)
	assert.Equal(t, []string{"ddd", "eee", "fff"}, actual.SDKKeys)
}

func assertRuntime(t *testing.T, actual config.RuntimeConfig) {
	assert.Equal(t, 1, actual.BlockProfileRate)
	assert.Equal(t, 2, actual.MutexProfileFraction)
}

func assertServer(t *testing.T, actual config.ServerConfig, assertPlugins bool) {
	assert.Equal(t, 5*time.Second, actual.ReadTimeout)
	assert.Equal(t, 10*time.Second, actual.WriteTimeout)
	assert.Equal(t, "/healthcheck", actual.HealthCheckPath)
	assert.Equal(t, "keyfile", actual.KeyFile)
	assert.Equal(t, "certfile", actual.CertFile)
	assert.Equal(t, []string{"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256", "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"}, actual.DisabledCiphers)
	assert.Equal(t, "1.2.3.4", actual.Host)
	assert.Equal(t, 100, actual.BatchRequests.OperationsLimit)
	assert.Equal(t, 5, actual.BatchRequests.MaxConcurrency)

	if assertPlugins {
		assert.Equal(t, config.PluginConfigs{"plugin": map[string]interface{}{}}, actual.Interceptors)
	}
}

func assertClient(t *testing.T, actual config.ClientConfig) {
	assert.Equal(t, 10*time.Second, actual.PollingInterval)
	assert.Equal(t, 1, actual.BatchSize)
	assert.Equal(t, 10, actual.QueueSize)
	assert.Equal(t, 1*time.Minute, actual.FlushInterval)
	assert.Equal(t, "https://localhost/v1/%s.json", actual.DatafileURLTemplate)
	assert.Equal(t, "https://logx.localhost.com/v1", actual.EventURL)
	assert.Equal(t, "custom-regex", actual.SdkKeyRegex)

	assert.Equal(t, "in-memory", actual.UserProfileService["default"])
	userProfileServices := map[string]interface{}{
		"in-memory": map[string]interface{}{
			// Viper.set is case in-sensitive
			"storagestrategy": "fifo",
		},
		"redis": map[string]interface{}{
			"host":     "localhost:6379",
			"password": "",
		},
		"rest": map[string]interface{}{
			"host":       "http://localhost",
			"lookuppath": "/ups/lookup",
			"savepath":   "/ups/save",
			"headers":    map[string]interface{}{"content-type": "application/json"},
			"async":      true,
		},
		"custom": map[string]interface{}{
			"path": "http://test2.com",
		},
	}
	assert.Equal(t, userProfileServices, actual.UserProfileService["services"])

	assert.Equal(t, "in-memory", actual.ODPCache["default"])
	odpCacheServices := map[string]interface{}{
		"redis": map[string]interface{}{
			"host":     "localhost:6379",
			"password": "",
		},
		"custom": map[string]interface{}{
			"path": "http://test2.com",
		},
	}
	actualCacheServices := actual.ODPCache["services"].(map[string]interface{})

	assert.Equal(t, odpCacheServices["redis"], actualCacheServices["redis"])
	assert.Equal(t, odpCacheServices["custom"], actualCacheServices["custom"])

	actualInMemoryService := actualCacheServices["in-memory"].(map[string]interface{})
	assert.EqualValues(t, 100, actualInMemoryService["size"])
	assert.EqualValues(t, 5, actualInMemoryService["timeout"])
}

func assertLog(t *testing.T, actual config.LogConfig) {
	assert.True(t, actual.Pretty)
	assert.False(t, actual.IncludeSDKKey)
	assert.Equal(t, "debug", actual.Level)
}

func assertAdmin(t *testing.T, actual config.AdminConfig) {
	assert.Equal(t, "3002", actual.Port)
}

func assertAdminAuth(t *testing.T, actual config.ServiceAuthConfig) {
	assert.Equal(t, 30*time.Minute, actual.TTL)
	assert.Len(t, actual.HMACSecrets, 2)
	assert.Equal(t, "efgh", actual.HMACSecrets[0])
	assert.Equal(t, "ijkl", actual.HMACSecrets[1])
	assert.Equal(t, config.OAuthClientCredentials{
		ID:         "clientid2",
		SecretHash: "clientsecret2",
		SDKKeys:    []string{"123"},
	}, actual.Clients[0])
	assert.Equal(t, "admin_jwks_url", actual.JwksURL)
	assert.Equal(t, 25*time.Second, actual.JwksUpdateInterval)
}

func assertAPI(t *testing.T, actual config.APIConfig) {
	assert.Equal(t, 100, actual.MaxConns)
	assert.Equal(t, "3000", actual.Port)
	assert.Equal(t, true, actual.EnableNotifications)
	assert.Equal(t, true, actual.EnableOverrides)
}

func assertAPIAuth(t *testing.T, actual config.ServiceAuthConfig) {
	assert.Equal(t, 30*time.Minute, actual.TTL)
	assert.Len(t, actual.HMACSecrets, 2)
	assert.Equal(t, "abcd", actual.HMACSecrets[0])
	assert.Equal(t, "efgh", actual.HMACSecrets[1])
	assert.Equal(t, config.OAuthClientCredentials{
		ID:         "clientid1",
		SecretHash: "clientsecret1",
		SDKKeys:    []string{"123"},
	}, actual.Clients[0])
	assert.Equal(t, "api_jwks_url", actual.JwksURL)
	assert.Equal(t, 25*time.Second, actual.JwksUpdateInterval)
}

func assertAPICORS(t *testing.T, actual config.CORSConfig) {
	assert.Equal(t, []string{"http://test1.com", "http://test2.com"}, actual.AllowedOrigins)
	assert.Equal(t, []string{"POST", "GET", "OPTIONS"}, actual.AllowedMethods)
	assert.Equal(t, []string{"Accept", "Authorization"}, actual.AllowedHeaders)
	assert.Equal(t, []string{"Header1"}, actual.ExposedHeaders)
	assert.Equal(t, true, actual.AllowedCredentials)
	assert.Equal(t, 500, actual.MaxAge)
}

func assertWebhook(t *testing.T, actual config.WebhookConfig) {
	assert.Equal(t, "3001", actual.Port)
	assert.Equal(t, "secret-10000", actual.Projects[10000].Secret)
	assert.Equal(t, []string{"aaa", "bbb", "ccc"}, actual.Projects[10000].SDKKeys)
	assert.True(t, actual.Projects[10000].SkipSignatureCheck)
	assert.Equal(t, "secret-20000", actual.Projects[20000].Secret)
	assert.Equal(t, []string{"xxx", "yyy", "zzz"}, actual.Projects[20000].SDKKeys)
	assert.False(t, actual.Projects[20000].SkipSignatureCheck)
}

func TestViperYaml(t *testing.T) {
	v := viper.New()
	v.Set("config.filename", "./testdata/default.yaml")

	err := initConfig(v)
	assert.NoError(t, err)

	actual := loadConfig(v)

	assertRoot(t, actual)
	assertServer(t, actual.Server, true)
	assertClient(t, actual.Client)
	assertLog(t, actual.Log)
	assertAdmin(t, actual.Admin)
	assertAdminAuth(t, actual.Admin.Auth)
	assertAPI(t, actual.API)
	assertAPIAuth(t, actual.API.Auth)
	assertAPICORS(t, actual.API.CORS)
	assertWebhook(t, actual.Webhook)
	assertRuntime(t, actual.Runtime)
}

func TestViperProps(t *testing.T) {
	v := viper.New()

	v.Set("version", "0.1.0")
	v.Set("author", "Optimizely Inc.")
	v.Set("name", "optimizely")
	v.Set("sdkkeys", []string{"ddd", "eee", "fff"})

	v.Set("server.readTimeout", 5*time.Second)
	v.Set("server.writeTimeout", 10*time.Second)
	v.Set("server.healthCheckPath", "/healthcheck")
	v.Set("server.certFile", "certfile")
	v.Set("server.keyFile", "keyfile")
	v.Set("server.disabledCiphers", "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384")
	v.Set("server.host", "1.2.3.4")
	v.Set("server.batchRequests.operationsLimit", "100")
	v.Set("server.batchRequests.maxConcurrency", "5")
	v.Set("server.interceptors", config.PluginConfigs{"plugin": map[string]interface{}{}})

	v.Set("client.pollingInterval", 10*time.Second)
	v.Set("client.batchSize", 1)
	v.Set("client.queueSize", 10)
	v.Set("client.flushInterval", 1*time.Minute)
	v.Set("client.datafileURLTemplate", "https://localhost/v1/%s.json")
	v.Set("client.eventURL", "https://logx.localhost.com/v1")
	v.Set("client.sdkKeyRegex", "custom-regex")
	upsServices := map[string]interface{}{
		"in-memory": map[string]interface{}{
			"storageStrategy": "fifo",
		},
		"redis": map[string]interface{}{
			"host":     "localhost:6379",
			"password": "",
		},
		"rest": map[string]interface{}{
			"host":       "http://localhost",
			"lookuppath": "/ups/lookup",
			"savepath":   "/ups/save",
			"headers":    map[string]interface{}{"content-type": "application/json"},
			"async":      true,
		},
		"custom": map[string]interface{}{
			"path": "http://test2.com",
		},
	}
	userProfileServices := map[string]interface{}{
		"default":  "in-memory",
		"services": upsServices,
	}
	v.Set("client.userProfileService", userProfileServices)

	odpCacheServices := map[string]interface{}{
		"in-memory": map[string]interface{}{
			"size":    100,
			"timeout": 5,
		},
		"redis": map[string]interface{}{
			"host":     "localhost:6379",
			"password": "",
		},
		"custom": map[string]interface{}{
			"path": "http://test2.com",
		},
	}
	odpCache := map[string]interface{}{
		"default":  "in-memory",
		"services": odpCacheServices,
	}
	v.Set("client.odpCache", odpCache)

	v.Set("log.pretty", true)
	v.Set("log.includeSdkKey", false)
	v.Set("log.level", "debug")

	v.Set("admin.port", "3002")
	v.Set("admin.auth.ttl", "30m")
	v.Set("admin.auth.hmacSecrets", "efgh,ijkl")
	v.Set("admin.auth.jwksURL", "admin_jwks_url")
	v.Set("admin.auth.jwksUpdateInterval", "25s")

	v.Set("admin.auth.clients", []map[string]interface{}{
		{
			"id":         "clientid2",
			"secretHash": "clientsecret2",
			"sdkKeys":    []string{"123"},
		},
	})

	v.Set("api.maxConns", 100)
	v.Set("api.enableNotifications", true)
	v.Set("api.enableOverrides", true)
	v.Set("api.port", "3000")
	v.Set("api.auth.ttl", "30m")

	v.Set("api.auth.hmacSecrets", "abcd,efgh")
	v.Set("api.auth.jwksURL", "api_jwks_url")
	v.Set("api.auth.jwksUpdateInterval", "25s")

	v.Set("api.auth.clients", []map[string]interface{}{
		{
			"id":         "clientid1",
			"secretHash": "clientsecret1",
			"sdkKeys":    []string{"123"},
		},
	})

	v.Set("webhook.port", "3001")
	v.Set("webhook.projects.10000.secret", "secret-10000")
	v.Set("webhook.projects.10000.sdkKeys", []string{"aaa", "bbb", "ccc"})
	v.Set("webhook.projects.10000.skipSignatureCheck", true)
	v.Set("webhook.projects.20000.secret", "secret-20000")
	v.Set("webhook.projects.20000.sdkKeys", []string{"xxx", "yyy", "zzz"})
	v.Set("webhook.projects.20000.skipSignatureCheck", false)

	v.Set("runtime.blockProfileRate", 1)
	v.Set("runtime.mutexProfileFraction", 2)

	assert.NoError(t, initConfig(v))
	actual := loadConfig(v)

	assertRoot(t, actual)
	assertServer(t, actual.Server, true)
	assertClient(t, actual.Client)
	assertLog(t, actual.Log)
	assertAdmin(t, actual.Admin)
	assertAdminAuth(t, actual.Admin.Auth)
	assertAPI(t, actual.API)
	assertAPIAuth(t, actual.API.Auth)
	assertWebhook(t, actual.Webhook)
	assertRuntime(t, actual.Runtime)
}

func TestViperEnv(t *testing.T) {
	_ = os.Setenv("OPTIMIZELY_VERSION", "0.1.0")
	_ = os.Setenv("OPTIMIZELY_AUTHOR", "Optimizely Inc.")
	_ = os.Setenv("OPTIMIZELY_NAME", "optimizely")
	_ = os.Setenv("OPTIMIZELY_SDKKEYS", "ddd,eee,fff")

	_ = os.Setenv("OPTIMIZELY_SERVER_READTIMEOUT", "5s")
	_ = os.Setenv("OPTIMIZELY_SERVER_WRITETIMEOUT", "10s")
	_ = os.Setenv("OPTIMIZELY_SERVER_HEALTHCHECKPATH", "/healthcheck")
	_ = os.Setenv("OPTIMIZELY_SERVER_CERTFILE", "certfile")
	_ = os.Setenv("OPTIMIZELY_SERVER_KEYFILE", "keyfile")
	_ = os.Setenv("OPTIMIZELY_SERVER_DISABLEDCIPHERS", "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384")
	_ = os.Setenv("OPTIMIZELY_SERVER_HOST", "1.2.3.4")
	_ = os.Setenv("OPTIMIZELY_SERVER_BATCHREQUESTS_MAXCONCURRENCY", "5")
	_ = os.Setenv("OPTIMIZELY_SERVER_BATCHREQUESTS_OPERATIONSLIMIT", "100")

	_ = os.Setenv("OPTIMIZELY_CLIENT_POLLINGINTERVAL", "10s")
	_ = os.Setenv("OPTIMIZELY_CLIENT_BATCHSIZE", "1")
	_ = os.Setenv("OPTIMIZELY_CLIENT_QUEUESIZE", "10")
	_ = os.Setenv("OPTIMIZELY_CLIENT_FLUSHINTERVAL", "1m")
	_ = os.Setenv("OPTIMIZELY_CLIENT_DATAFILEURLTEMPLATE", "https://localhost/v1/%s.json")
	_ = os.Setenv("OPTIMIZELY_CLIENT_EVENTURL", "https://logx.localhost.com/v1")
	_ = os.Setenv("OPTIMIZELY_CLIENT_SDKKEYREGEX", "custom-regex")
	_ = os.Setenv("OPTIMIZELY_CLIENT_USERPROFILESERVICE", `{"default":"in-memory","services":{"in-memory":{"storagestrategy":"fifo"},"redis":{"host":"localhost:6379","password":""},"rest":{"host":"http://localhost","lookuppath":"/ups/lookup","savepath":"/ups/save","headers":{"content-type":"application/json"},"async":true},"custom":{"path":"http://test2.com"}}}`)
	_ = os.Setenv("OPTIMIZELY_CLIENT_ODPCACHE", `{"default":"in-memory","services":{"in-memory":{"size":100,"timeout":5},"redis":{"host":"localhost:6379","password":""},"custom":{"path":"http://test2.com"}}}`)

	_ = os.Setenv("OPTIMIZELY_LOG_PRETTY", "true")
	_ = os.Setenv("OPTIMIZELY_LOG_INCLUDESDKKEY", "false")
	_ = os.Setenv("OPTIMIZELY_LOG_LEVEL", "debug")

	_ = os.Setenv("OPTIMIZELY_ADMIN_PORT", "3002")

	_ = os.Setenv("OPTIMIZELY_API_MAXCONNS", "100")
	_ = os.Setenv("OPTIMIZELY_API_PORT", "3000")
	_ = os.Setenv("OPTIMIZELY_API_ENABLENOTIFICATIONS", "true")
	_ = os.Setenv("OPTIMIZELY_API_ENABLEOVERRIDES", "true")

	_ = os.Setenv("OPTIMIZELY_WEBHOOK_PORT", "3001")
	_ = os.Setenv("OPTIMIZELY_WEBHOOK_PROJECTS_10000_SECRET", "secret-10000")
	_ = os.Setenv("OPTIMIZELY_WEBHOOK_PROJECTS_10000_SDKKEYS", "aaa,bbb,ccc")
	_ = os.Setenv("OPTIMIZELY_WEBHOOK_PROJECTS_10000_SKIPSIGNATURECHECK", "true")
	_ = os.Setenv("OPTIMIZELY_WEBHOOK_PROJECTS_20000_SECRET", "secret-20000")
	_ = os.Setenv("OPTIMIZELY_WEBHOOK_PROJECTS_20000_SDKKEYS", "xxx,yyy,zzz")
	_ = os.Setenv("OPTIMIZELY_WEBHOOK_PROJECTS_20000_SKIPSIGNATURECHECK", "false")

	_ = os.Setenv("OPTIMIZELY_RUNTIME_BLOCKPROFILERATE", "1")
	_ = os.Setenv("OPTIMIZELY_RUNTIME_MUTEXPROFILEFRACTION", "2")

	v := viper.New()
	assert.NoError(t, initConfig(v))
	actual := loadConfig(v)

	assertRoot(t, actual)
	assertServer(t, actual.Server, false)
	assertClient(t, actual.Client)
	assertLog(t, actual.Log)
	assertAdmin(t, actual.Admin)
	assertAPI(t, actual.API)
	//assertWebhook(t, actual.Webhook) // Maps don't appear to be supported
	assertRuntime(t, actual.Runtime)
}

func TestLoggingWithIncludeSdkKey(t *testing.T) {
	// Test default IncludeSDKKey value
	assert.True(t, optimizely.ShouldIncludeSDKKey)
	// Test log config should reflect on optimizely.ShouldIncludeSDKKey
	initLogging(config.LogConfig{
		IncludeSDKKey: false,
	})
	assert.False(t, optimizely.ShouldIncludeSDKKey)
}
