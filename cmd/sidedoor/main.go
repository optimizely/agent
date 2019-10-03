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
	"sync"

	"github.com/optimizely/sidedoor/pkg/api"
	"github.com/optimizely/sidedoor/pkg/webhook"
)

func main() {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		log.Printf("Optimizely API server started")
		apiRouter := api.NewDefaultRouter()
		log.Fatal(http.ListenAndServe(":8080", apiRouter))
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		// TODO optionally not start this if user is not interested in webhooks
		log.Printf("Optimizely webhook server started")
		webhookRouter := webhook.NewRouter()
		log.Fatal(http.ListenAndServe(":8085", webhookRouter))
		wg.Done()
	}()

	wg.Wait()
}
