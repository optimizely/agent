/****************************************************************************
 * Copyright 2019, Optimizely, Inc. and contributors                        *
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
	"bytes"
	"context"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"gopkg.in/yaml.v2"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/optimizely/agent/config"
	"github.com/optimizely/agent/pkg/metrics"
	"github.com/optimizely/agent/pkg/optimizely"
	"github.com/optimizely/agent/pkg/routers"
	"github.com/optimizely/agent/pkg/server"
)

// Version holds the admin version
var Version string // default set at compile time

func initConfig(v *viper.Viper) error {
	// Set explicit defaults
	v.SetDefault("config.filename", "config.yaml") // Configuration file name

	// Load defaults from the AgentConfig by loading the marshaled values as yaml
	// https://github.com/spf13/viper/issues/188
	defaultConf := config.NewDefaultConfig()
	defaultConf.Version = Version
	b, err := yaml.Marshal(defaultConf)
	if err != nil {
		return err
	}

	dc := bytes.NewReader(b)
	v.SetConfigType("yaml")
	return v.MergeConfig(dc)
}

func loadConfig(v *viper.Viper) *config.AgentConfig {
	// Configure environment variables
	v.SetEnvPrefix("optimizely")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read configuration from file
	configFile := v.GetString("config.filename")
	v.SetConfigFile(configFile)
	if err := v.MergeInConfig(); err != nil {
		log.Info().Err(err).Msg("Skip loading configuration from config file.")
	}

	conf := &config.AgentConfig{}
	if err := v.Unmarshal(conf); err != nil {
		log.Info().Err(err).Msg("Unable to marshal configuration.")
	}

	// https://github.com/spf13/viper/issues/406
	if plugins, ok := v.Get("server.plugins").(map[string]interface{}); ok {
		conf.Server.Plugins = plugins
	}

	return conf
}

func initLogging(conf config.LogConfig) {
	if conf.Pretty {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	if lvl, err := zerolog.ParseLevel(conf.Level); err != nil {
		log.Warn().Err(err).Msg("Error parsing log level")
	} else {
		log.Logger = log.Logger.Level(lvl)
	}
}

func setRuntimeEnvironment(conf config.RuntimeConfig) {
	if conf.BlockProfileRate != 0 {
		log.Warn().Msgf("Setting non-zero blockProfileRate is NOT recommended for production")
		runtime.SetBlockProfileRate(conf.BlockProfileRate)
	}

	if conf.MutexProfileFraction != 0 {
		log.Warn().Msgf("Setting non-zero mutexProfileFraction is NOT recommended for production")
		runtime.SetMutexProfileFraction(conf.MutexProfileFraction)
	}
}

func main() {
	v := viper.New()
	if err := initConfig(v); err != nil {
		log.Panic().Err(err).Msg("Unable to initialize config")
	}

	conf := loadConfig(v)
	initLogging(conf.Log)

	conf.LogConfigWarnings()

	setRuntimeEnvironment(conf.Runtime)

	agentMetricsRegistry := metrics.NewRegistry()
	sdkMetricsRegistry := optimizely.NewRegistry(agentMetricsRegistry)

	ctx, cancel := context.WithCancel(context.Background()) // Create default service context
	sg := server.NewGroup(ctx, conf.Server)                 // Create a new server group to manage the individual http listeners
	optlyCache := optimizely.NewCache(ctx, conf.Client, sdkMetricsRegistry)
	optlyCache.Init(conf.SDKKeys)

	// goroutine to check for signals to gracefully shutdown listeners
	go func() {
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

		// Wait for signal
		sig := <-signalChannel
		log.Info().Msgf("Received signal: %s\n", sig)
		cancel()
	}()

	apiRouter := routers.NewDefaultAPIRouter(optlyCache, conf.API, agentMetricsRegistry)
	adminRouter := routers.NewAdminRouter(*conf)

	log.Info().Str("version", conf.Version).Msg("Starting services.")
	sg.GoListenAndServe("api", conf.API.Port, apiRouter)
	sg.GoListenAndServe("webhook", conf.Webhook.Port, routers.NewWebhookRouter(optlyCache, conf.Webhook))
	sg.GoListenAndServe("admin", conf.Admin.Port, adminRouter) // Admin should be added last.

	// wait for server group to shutdown
	if err := sg.Wait(); err == nil {
		log.Info().Msg("Exiting.")
	} else {
		log.Fatal().Err(err).Msg("Exiting.")
	}

	optlyCache.Wait()
}
