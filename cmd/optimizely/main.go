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
	"gopkg.in/yaml.v2"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/optimizely/sidedoor/config"
	"github.com/optimizely/sidedoor/pkg/admin"
	"github.com/optimizely/sidedoor/pkg/api"
	"github.com/optimizely/sidedoor/pkg/optimizely"
	"github.com/optimizely/sidedoor/pkg/server"
	"github.com/optimizely/sidedoor/pkg/webhook"
)

// Version holds the admin version
var Version string // default set at compile time

//https://github.com/spf13/viper/issues/188
func loadConfig(v *viper.Viper) (*config.AgentConfig, error) {
	// Set defaults
	v.SetDefault("config.filename", "config.yaml") // Configuration file name
	v.SetDefault("app.version", Version)           // Application version

	defaultConf := config.NewAgentConfig()
	b, err := yaml.Marshal(defaultConf)
	if err != nil {
		return defaultConf, err
	}

	dc := bytes.NewReader(b)
	v.SetConfigType("yaml")
	if err2 := v.MergeConfig(dc); err2 != nil {
		return defaultConf, err2
	}

	// Read configuration from file
	configFile := v.GetString("config.filename")
	v.SetConfigFile(configFile)
	if err3 := v.MergeInConfig(); err3 != nil {
		if _, ok := err3.(viper.ConfigParseError); ok {
			return defaultConf, err3
		}
		// dont return error if file is missing. overwrite file is optional
	}

	// Configure environment variables
	v.SetEnvPrefix("optimizely")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetConfigType("yaml")

	// refresh configuration with all merged values
	conf := &config.AgentConfig{}
	err = v.Unmarshal(&conf)
	return conf, err
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

func main() {

	conf, err := loadConfig(viper.New())
	initLogging(conf.Log)

	if err != nil {
		log.Info().Err(err).Msg("Skip loading configuration from config file.")
	}

	log.Info().Str("version", viper.GetString("app.version")).Msg("Starting services.")
	log.Info().Str("api-port", viper.GetString("api.port")).Msg("Starting services.")

	ctx, cancel := context.WithCancel(context.Background()) // Create default service context
	sg := server.NewGroup(ctx, conf.Server)                 // Create a new server group to manage the individual http listeners
	optlyCache := optimizely.NewCache(ctx, conf.Optly)

	// goroutine to check for signals to gracefully shutdown listeners
	go func() {
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

		// Wait for signal
		sig := <-signalChannel
		log.Info().Msgf("Received signal: %s\n", sig)
		cancel()
	}()

	sg.GoListenAndServe("api", conf.API.Port, api.NewDefaultRouter(optlyCache, conf.API))
	sg.GoListenAndServe("webhook", conf.Webhook.Port, webhook.NewRouter(optlyCache, conf.Webhook))
	sg.GoListenAndServe("admin", conf.Admin.Port, admin.NewRouter(conf.Admin)) // Admin should be added last.

	// wait for server group to shutdown
	if err := sg.Wait(); err == nil {
		log.Info().Msg("Exiting.")
	} else {
		log.Fatal().Err(err).Msg("Exiting.")
	}

	optlyCache.Wait()
}
