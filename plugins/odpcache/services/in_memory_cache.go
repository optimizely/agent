/****************************************************************************
 * Copyright 2023, Optimizely, Inc. and contributors                        *
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

// Package services //
package services

import (
	"time"

	"github.com/optimizely/agent/plugins/odpcache"
	"github.com/optimizely/go-sdk/pkg/odp/cache"
)

// InMemoryCache represents the in-memory implementation of Cache interface
type InMemoryCache struct {
	Size    int `json:"size"`
	Timeout int `json:"timeout"`
	*cache.LRUCache
}

// Lookup is used to retrieve segments
func (i *InMemoryCache) Lookup(key string) (segments interface{}) {
	if i.LRUCache == nil {
		i.initClient()
		return
	}
	return i.LRUCache.Lookup(key)
}

// Save is used to save segments
func (i *InMemoryCache) Save(key string, value interface{}) {
	if i.LRUCache == nil {
		i.initClient()
	}
	i.LRUCache.Save(key, value)
}

// Reset is used to reset segments
func (i *InMemoryCache) Reset() {
	if i.LRUCache != nil {
		i.LRUCache.Reset()
	}
}

func (i *InMemoryCache) initClient() {
	i.LRUCache = cache.NewLRUCache(i.Size, time.Duration(i.Timeout*int(time.Second)))
}

func init() {
	inMemoryCacheCreator := func() cache.Cache {
		return &InMemoryCache{}
	}
	odpcache.Add("in-memory", inMemoryCacheCreator)
}
