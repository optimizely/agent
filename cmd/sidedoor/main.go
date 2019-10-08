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
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"net/http"
	"strings"
	"sync"

	"github.com/optimizely/sidedoor/pkg/api"
	"github.com/optimizely/sidedoor/pkg/webhook"
)

func init() {
	loadConfig()
}

func loadConfig() {
	viper.SetEnvPrefix("sidedoor")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set config file
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Info().Msg("No config file found.")
	}

	viper.AutomaticEnv()

	// Port for serving Optimizely APIs
	viper.SetDefault("api.port", "8080")
	// Property to turn webhook service on/off
	viper.SetDefault("webhook.enabled", false)
	// Port for webhook service
	viper.SetDefault("webhook.port", "8085")
}

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		apiRouter := api.NewDefaultRouter()
		apiPort := viper.GetString("api.port")
		log.Printf("Optimizely API server started at port " + apiPort)
		if err := http.ListenAndServe(":" + apiPort, apiRouter); err != nil {
			log.Fatal().Err(err).Msg("Failed to start Optimizely API server.")
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		// TODO optionally not start this if user is not interested in webhooks
		webhookEnabled := viper.GetBool("webhook.enabled")
		if !webhookEnabled {
			log.Printf("Webhook service opted out.")
			return
		}
		webhookRouter := webhook.NewRouter()
		webhookPort := viper.GetString("webhook.port")
		log.Printf("Optimizely webhook server started at port " + webhookPort)
		if err := http.ListenAndServe(":" + webhookPort, webhookRouter); err != nil {
			log.Fatal().Err(err).Msg("Failed to start Optimizely webhook server.")
		}
		wg.Done()
	}()

	wg.Wait()
}
