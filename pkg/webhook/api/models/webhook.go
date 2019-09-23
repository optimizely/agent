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

// Package models //
package models

// DatafileUpdateData model which represents data specific to datafile update
type DatafileUpdateData struct {
    Revision       int32    `json:"revision"`
    OriginURL      string   `json:"origin_url"`
    CDNUrl         string   `json:"cdn_url"`
    Environment    string   `json:"environment"`
}

// WebhookMessage model which represents any message received from Optimizely
type WebhookMessage struct {
    ProjectId    int64					`json:"project_id"`
    Timestamp    int64					`json:"timestamp"`
    Event        string					`json:"event"`
    Data         DatafileUpdateData     `json:"data"`
}
