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
	"log"
	"net/http"
	"strings"
	"sync"
	"github.com/spf13/viper"

	"github.com/optimizely/sidedoor/pkg/api"
	"github.com/optimizely/sidedoor/pkg/webhook"
)

func loadConfig() {
	viper.SetEnvPrefix("SIDEDOOR")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Port for serving Optimizely APIs
	viper.SetDefault("api.port", "8080")
	// Property to turn webhook service on/off
	viper.SetDefault("webhook.enabled", false)
	// Port for webhook service
	viper.SetDefault("webhook.port", "8085")
	// Path to file for configuring webhooks
	viper.SetDefault("webhook.config", "config.yaml")
}


func main() {
	loadConfig()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		apiRouter := api.NewDefaultRouter()
		apiPort := viper.GetString("api.port")
		log.Printf("Optimizely API server started at port " + apiPort)
		log.Fatal(http.ListenAndServe(":" + apiPort, apiRouter))
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
		log.Fatal(http.ListenAndServe(":" + webhookPort, webhookRouter))
		wg.Done()
	}()

	wg.Wait()
}
