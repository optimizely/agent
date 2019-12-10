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
	"time"

	"github.com/optimizely/sidedoor/pkg/admin"
	"github.com/optimizely/sidedoor/pkg/api"
	"github.com/optimizely/sidedoor/pkg/optimizely"
	"github.com/optimizely/sidedoor/pkg/server"
	"github.com/optimizely/sidedoor/pkg/webhook"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Version holds the admin version
var Version string // default set at compile time

func loadConfig() error {
	// Set defaults
	viper.SetDefault("config.filename", "config.yaml") // Configuration file name

	viper.SetDefault("app.version", Version)     // Application version
	viper.SetDefault("app.author", "Optimizely") // Application author
	viper.SetDefault("app.name", "Sidedoor")     // Appplication name

	viper.SetDefault("api.enabled", true) // Property to turn api service on/off
	viper.SetDefault("api.port", "8080")  // Port for serving Optimizely APIs

	viper.SetDefault("webhook.enabled", true) // Property to turn webhook service on/off
	viper.SetDefault("webhook.port", "8085")  // Port for webhook service

	viper.SetDefault("admin.enabled", true)
	viper.SetDefault("admin.port", "8088") // Port for admin service

	viper.SetDefault("log.level", "info") // Set default log level

	viper.SetDefault("server.readtimeout", 5*time.Second)
	viper.SetDefault("server.writetimeout", 10*time.Second)

	viper.SetDefault("metrics.pollingfreqency", time.Minute) // Set default polling frequency for metric collection

	// Configure environment variables
	viper.SetEnvPrefix("sidedoor")
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

	log.Info().Str("version", viper.GetString("app.version")).Msg("Starting services.")

	ctx := context.Background()         // Create default service context
	sg := server.NewGroup(ctx)          // Create a new server group to manage the individual http listeners
	optlyCache := optimizely.NewCache() // TODO pass ctx

	// goroutine to check for signals to gracefully shutdown listeners
	go func() {
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

		// Wait for signal
		sig := <-signalChannel
		log.Info().Msgf("Received signal: %s\n", sig)
		sg.Shutdown()
	}()

	sg.GoListenAndServe("api", api.NewDefaultRouter(optlyCache))
	sg.GoListenAndServe("webhook", webhook.NewDefaultRouter(optlyCache))
	sg.GoListenAndServe("admin", admin.NewRouter()) // Admin should be added last.

	// wait for server group to shutdown
	if err := sg.Wait(); err == nil {
		log.Info().Msg("Exiting.")
	} else {
		log.Fatal().Err(err).Msg("Exiting.")
	}
}
