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
	"context"
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

func loadConfig() error {
	// Set defaults
	viper.SetDefault("config.filename", "config.yaml") // Configuration file name
	viper.SetDefault("app.version", Version)           // Application version

	// Configure environment variables
	viper.SetEnvPrefix("optimizely")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Read configuration from file
	configFile := viper.GetString("config.filename")
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")
	return viper.ReadInConfig()
}

func initLogging() {
	if viper.GetBool("log.pretty") {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	if lvl, err := zerolog.ParseLevel(viper.GetString("log.level")); err != nil {
		log.Warn().Err(err).Msg("Error parsing log level")
	} else {
		log.Logger = log.Logger.Level(lvl)
	}
}

func main() {

	err := loadConfig()
	initLogging()

	if err != nil {
		log.Info().Err(err).Msg("Skip loading configuration from config file.")
	}

	conf := config.NewAgentConfig()
	if err := viper.Unmarshal(conf); err != nil {
		log.Error().Err(err).Msg("Unable to marshall configuration")
	}

	log.Info().Str("version", viper.GetString("app.version")).Msg("Starting services.")

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
