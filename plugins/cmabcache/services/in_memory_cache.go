/****************************************************************************
 * Copyright 2025, Optimizely, Inc. and contributors                        *
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
	"github.com/optimizely/agent/plugins/cmabcache"
	"github.com/optimizely/agent/plugins/utils"
	"github.com/optimizely/go-sdk/v2/pkg/cache"
)

// InMemoryCache represents the in-memory implementation of Cache interface
type InMemoryCache struct {
	Size    int            `json:"size"`
	Timeout utils.Duration `json:"timeout"`
	*cache.LRUCache
}

// Lookup is used to retrieve cached CMAB decisions
func (i *InMemoryCache) Lookup(key string) interface{} {
	if i.LRUCache == nil {
		i.initClient()
		return nil
	}
	return i.LRUCache.Lookup(key)
}

// Save is used to save CMAB decisions
func (i *InMemoryCache) Save(key string, value interface{}) {
	if i.LRUCache == nil {
		i.initClient()
	}
	i.LRUCache.Save(key, value)
}

// Remove is used to remove a specific CMAB decision from cache
func (i *InMemoryCache) Remove(key string) {
	if i.LRUCache == nil {
		return
	}
	i.LRUCache.Remove(key)
}

// Reset is used to reset all CMAB decisions
func (i *InMemoryCache) Reset() {
	if i.LRUCache != nil {
		i.LRUCache.Reset()
	}
}

func (i *InMemoryCache) initClient() {
	i.LRUCache = cache.NewLRUCache(i.Size, i.Timeout.Duration)
}

func init() {
	inMemoryCacheCreator := func() cache.Cache {
		return &InMemoryCache{}
	}
	cmabcache.Add("in-memory", inMemoryCacheCreator)
}
