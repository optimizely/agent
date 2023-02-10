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
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

type InMemoryCacheTestSuite struct {
	suite.Suite
	cache InMemoryCache
}

func (im *InMemoryCacheTestSuite) SetupTest() {
	// To check if lifo is used by default
	im.cache = InMemoryCache{
		Capacity: 10,
	}
}

func (im *InMemoryCacheTestSuite) TestConcurrentSaveAndLookupFifo() {
	wg := sync.WaitGroup{}
	save := func(key string) {
		im.cache.Save(key, []string{key})
		wg.Done()
	}

	lookUp := func(key string) {
		expected := []string{key}
		actual := im.cache.Lookup(key)
		im.Equal(expected, actual)
		wg.Done()
	}

	// Save concurrently
	wg.Add(10)
	i := 1
	for i <= 10 {
		i++
		go save(strconv.Itoa(i))
	}
	wg.Wait()

	// Lookup and save concurrently
	wg.Add(20)
	i = 1
	for i <= 10 {
		i++
		go save(strconv.Itoa(i))
		go lookUp(strconv.Itoa(i))
	}
	wg.Wait()
	im.Equal(im.cache.Capacity, len(im.cache.SegmentsMap))
}

func (im *InMemoryCacheTestSuite) TestConcurrentSaveAndLookupLifo() {
	im.cache = InMemoryCache{
		StorageStrategy: "lifo",
		Capacity:        10,
	}
	wg := sync.WaitGroup{}
	save := func(key string) {
		im.cache.Save(key, []string{key})
		wg.Done()
	}

	lookUp := func(key string) {
		expected := []string{key}
		actual := im.cache.Lookup(key)
		im.Equal(expected, actual)
		wg.Done()
	}

	// Save concurrently
	wg.Add(10)
	i := 1
	for i <= 10 {
		i++
		go save(strconv.Itoa(i))
	}
	wg.Wait()

	// Lookup and save concurrently
	wg.Add(20)
	i = 1
	for i <= 10 {
		i++
		go save(strconv.Itoa(i))
		go lookUp(strconv.Itoa(i))
	}
	wg.Wait()
	im.Equal(im.cache.Capacity, len(im.cache.SegmentsMap))
}

func (im *InMemoryCacheTestSuite) TestOverrideFifo() {
	i := 1
	for i < 3 {
		strValue := strconv.Itoa(i)
		im.cache.Save("1", []string{strValue})
		i++
	}

	strValue := strconv.Itoa(2)
	expected := []string{strValue}
	actual := im.cache.Lookup("1")
	im.Equal(expected, actual)
}

func (im *InMemoryCacheTestSuite) TestOverrideLifo() {
	im.cache = InMemoryCache{
		StorageStrategy: "lifo",
		Capacity:        10,
	}
	i := 1
	for i < 3 {
		strValue := strconv.Itoa(i)
		im.cache.Save("1", []string{strValue})
		i++
	}

	strValue := strconv.Itoa(2)
	expected := []string{strValue}
	actual := im.cache.Lookup("1")
	im.Equal(expected, actual)
}

func (im *InMemoryCacheTestSuite) TestSaveEmptySegmentsFifo() {
	strValue := strconv.Itoa(1)
	im.cache.Save(strValue, []string{strValue})

	// Save empty segments
	im.cache.Save(strValue, []string{})

	actual := im.cache.Lookup(strValue)
	im.Equal([]string{}, actual)
}

func (im *InMemoryCacheTestSuite) TestSaveEmptySegmentsLifo() {
	im.cache = InMemoryCache{
		StorageStrategy: "lifo",
		Capacity:        10,
	}
	strValue := strconv.Itoa(1)
	im.cache.Save(strValue, []string{strValue})

	// Save empty segments
	im.cache.Save(strValue, []string{})

	actual := im.cache.Lookup(strValue)
	im.Equal([]string{}, actual)
}

func (im *InMemoryCacheTestSuite) TestCapacityFifoOrEmpty() {
	// Save 10 segments as capacity is given as 10
	i := 1
	for i <= 10 {
		i++
		strValue := strconv.Itoa(i)
		im.cache.Save(strValue, []string{strValue})
	}

	// Check all 10 segments were saved
	i = 1
	for i <= 10 {
		i++
		strValue := strconv.Itoa(i)
		expected := []string{strValue}
		actual := im.cache.Lookup(strValue)
		im.Equal(expected, actual)
	}

	// Save 3 more segments than the capacity
	i = 11
	for i <= 13 {
		i++
		strValue := strconv.Itoa(i)
		im.cache.Save(strValue, []string{strValue})
	}

	// Check first 3 segments were overwritten by newer 3 segments, total count still remains 10
	i = 4
	for i <= 13 {
		i++
		strValue := strconv.Itoa(i)
		expected := []string{strValue}
		actual := im.cache.Lookup(strValue)
		im.Equal(expected, actual)
	}
	im.Equal(10, len(im.cache.SegmentsMap))

	im.cache.Reset()
	im.Equal(0, len(im.cache.SegmentsMap))
	im.NotNil(im.cache.fifoOrderedSegments)
	im.Nil(im.cache.lifoOrderedSegments)
}

func (im *InMemoryCacheTestSuite) TestCapacityLifoOrEmpty() {
	im.cache = InMemoryCache{
		StorageStrategy: "lifo",
		Capacity:        10,
	}
	// Save 10 segments as capacity is given as 10
	i := 1
	for i <= 10 {
		strValue := strconv.Itoa(i)
		im.cache.Save(strValue, []string{strValue})
		i++
	}

	// Check all 10 segments were saved
	i = 1
	for i <= 10 {
		strValue := strconv.Itoa(i)
		expected := []string{strValue}
		actual := im.cache.Lookup(strValue)
		im.Equal(expected, actual)
		i++
	}

	// Save 3 more segments than the capacity
	i = 11
	for i <= 13 {
		strValue := strconv.Itoa(i)
		im.cache.Save(strValue, []string{strValue})
		i++
	}

	// Check latest entry was always overwritten by the newer entry
	values := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 13}
	for _, v := range values {
		strValue := strconv.Itoa(v)
		expected := []string{strValue}
		actual := im.cache.Lookup(strValue)
		im.Equal(expected, actual)
	}
	im.Equal(10, len(im.cache.SegmentsMap))

	im.cache.Reset()
	im.Equal(0, len(im.cache.SegmentsMap))
	im.NotNil(im.cache.lifoOrderedSegments)
	im.Nil(im.cache.fifoOrderedSegments)
}

func (im *InMemoryCacheTestSuite) TestZeroCapacityFifoOrEmpty() {
	im.cache = InMemoryCache{
		Capacity: 0,
	}
	// Save 200 segments as capacity is given as 0
	i := 1
	for i <= 200 {
		i++
		strValue := strconv.Itoa(i)
		im.cache.Save(strValue, []string{strValue})
	}

	// Check all 200 segments were saved
	i = 1
	for i <= 200 {
		i++
		strValue := strconv.Itoa(i)
		expected := []string{strValue}
		actual := im.cache.Lookup(strValue)
		im.Equal(expected, actual)
	}
	im.Equal(200, len(im.cache.SegmentsMap))
	im.Nil(im.cache.fifoOrderedSegments)

	im.cache.Reset()
	im.Equal(0, len(im.cache.SegmentsMap))
	im.Nil(im.cache.fifoOrderedSegments)
}

func (im *InMemoryCacheTestSuite) TestZeroCapacityLifo() {
	im.cache = InMemoryCache{
		StorageStrategy: "lifo",
		Capacity:        0,
	}
	// Save 200 segments as capacity is given as 0
	i := 1
	for i <= 200 {
		i++
		strValue := strconv.Itoa(i)
		im.cache.Save(strValue, []string{strValue})
	}

	// Check all 200 segments were saved
	i = 1
	for i <= 200 {
		i++
		strValue := strconv.Itoa(i)
		expected := []string{strValue}
		actual := im.cache.Lookup(strValue)
		im.Equal(expected, actual)
	}
	im.Equal(200, len(im.cache.SegmentsMap))
	im.Nil(im.cache.lifoOrderedSegments)

	im.cache.Reset()
	im.Equal(0, len(im.cache.SegmentsMap))
	im.Nil(im.cache.lifoOrderedSegments)
}

func TestInMemoryCacheTestSuite(t *testing.T) {
	suite.Run(t, new(InMemoryCacheTestSuite))
}
