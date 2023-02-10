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
	"sync"

	"github.com/optimizely/agent/plugins/odpcache"
	"github.com/optimizely/go-sdk/pkg/odp/cache"
)

// InMemoryCache represents the in-memory implementation of Cache interface
type InMemoryCache struct {
	Capacity int `json:"capacity"`
	// StorageStrategy defines the storage strategy. Supported values include fifo and lifo.
	StorageStrategy     string `json:"storageStrategy"`
	SegmentsMap         map[string]interface{}
	fifoOrderedSegments chan string
	lifoOrderedSegments []string
	lock                sync.RWMutex
	isReady             bool
}

// Lookup is used to retrieve segments
func (r *InMemoryCache) Lookup(key string) (segments interface{}) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	// Check if Cache is ready
	if !r.isReady {
		return
	}

	if s, ok := r.SegmentsMap[key]; ok {
		segments = s
	}
	return segments
}

// Save is used to save segments
func (r *InMemoryCache) Save(key string, value interface{}) {
	if key == "" {
		return
	}
	r.lock.Lock()
	defer r.lock.Unlock()

	// initialize properties if cache not ready
	if !r.isReady {
		r.reinitializeCache()
		r.isReady = true
	}

	// check if segments do not exist already
	if _, ok := r.SegmentsMap[key]; !ok {
		if r.Capacity > 0 {
			// Check if capacity has reached, if so, pop the entry from ordered list and map
			if len(r.SegmentsMap) == r.Capacity {
				var oldSegments string
				// pop entry from ordered list
				switch r.StorageStrategy {
				case "lifo":
					n := len(r.lifoOrderedSegments) - 1
					oldSegments = r.lifoOrderedSegments[n]
					r.lifoOrderedSegments[n] = "" // Erase element (write zero value)
					r.lifoOrderedSegments = r.lifoOrderedSegments[:n]
				default:
					// fifo by default
					oldSegments = <-r.fifoOrderedSegments
				}
				// remove entry from map
				delete(r.SegmentsMap, oldSegments)
			}

			// Push new entry to ordered list
			switch r.StorageStrategy {
			case "lifo":
				r.lifoOrderedSegments = append(r.lifoOrderedSegments, key)
			default:
				// fifo by default
				r.fifoOrderedSegments <- key
			}
		}
	}
	// Save new segments to map
	r.SegmentsMap[key] = value
}

// Reset is used to reset segments
func (r *InMemoryCache) Reset() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.reinitializeCache()
}

func (r *InMemoryCache) reinitializeCache() {
	// Initialize with capacity only if required
	if r.Capacity > 0 {
		switch r.StorageStrategy {
		case "lifo":
			r.lifoOrderedSegments = []string{}
		default:
			// fifo by default
			r.fifoOrderedSegments = make(chan string, r.Capacity)
		}
		r.SegmentsMap = make(map[string]interface{}, r.Capacity)
	} else {
		r.SegmentsMap = map[string]interface{}{}
	}
}

func init() {
	inMemoryCacheCreator := func() cache.Cache {
		return &InMemoryCache{}
	}
	odpcache.Add("in-memory", inMemoryCacheCreator)
}
