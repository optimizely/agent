/****************************************************************************
 * Copyright 2021, Optimizely, Inc. and contributors                        *
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

	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/stretchr/testify/suite"
)

type InMemoryUPSTestSuite struct {
	suite.Suite
	ups InMemoryUserProfileService
}

func (im *InMemoryUPSTestSuite) SetupTest() {
	// To check if lifo is used by default
	im.ups = InMemoryUserProfileService{
		Capacity: 10,
	}
}

func (im *InMemoryUPSTestSuite) TestConcurrentSaveAndLookupFifo() {
	wg := sync.WaitGroup{}
	saveProfile := func(counter string) {
		profile := decision.UserProfile{
			ID: counter,
			ExperimentBucketMap: map[decision.UserDecisionKey]string{
				decision.NewUserDecisionKey(counter): counter,
			},
		}
		im.ups.Save(profile)
		wg.Done()
	}

	lookUp := func(counter string) {
		expected := decision.UserProfile{
			ID: counter,
			ExperimentBucketMap: map[decision.UserDecisionKey]string{
				decision.NewUserDecisionKey(counter): counter,
			},
		}
		actual := im.ups.Lookup(counter)
		im.Equal(expected, actual)
		wg.Done()
	}

	// Save concurrently
	wg.Add(10)
	i := 1
	for i <= 10 {
		i++
		go saveProfile(strconv.Itoa(i))
	}
	wg.Wait()

	// Lookup and save concurrently
	wg.Add(20)
	i = 1
	for i <= 10 {
		i++
		go saveProfile(strconv.Itoa(i))
		go lookUp(strconv.Itoa(i))
	}
	wg.Wait()
	im.Equal(im.ups.Capacity, len(im.ups.ProfilesMap))
}

func (im *InMemoryUPSTestSuite) TestConcurrentSaveAndLookupLifo() {
	im.ups = InMemoryUserProfileService{
		Order:    "lifo",
		Capacity: 10,
	}
	wg := sync.WaitGroup{}
	saveProfile := func(counter string) {
		profile := decision.UserProfile{
			ID: counter,
			ExperimentBucketMap: map[decision.UserDecisionKey]string{
				decision.NewUserDecisionKey(counter): counter,
			},
		}
		im.ups.Save(profile)
		wg.Done()
	}

	lookUp := func(counter string) {
		expected := decision.UserProfile{
			ID: counter,
			ExperimentBucketMap: map[decision.UserDecisionKey]string{
				decision.NewUserDecisionKey(counter): counter,
			},
		}
		actual := im.ups.Lookup(counter)
		im.Equal(expected, actual)
		wg.Done()
	}

	// Save concurrently
	wg.Add(10)
	i := 1
	for i <= 10 {
		i++
		go saveProfile(strconv.Itoa(i))
	}
	wg.Wait()

	// Lookup and save concurrently
	wg.Add(20)
	i = 1
	for i <= 10 {
		i++
		go saveProfile(strconv.Itoa(i))
		go lookUp(strconv.Itoa(i))
	}
	wg.Wait()
	im.Equal(im.ups.Capacity, len(im.ups.ProfilesMap))
}

func (im *InMemoryUPSTestSuite) TestOverrideFifo() {
	i := 1
	for i < 3 {
		strValue := strconv.Itoa(i)
		profile := decision.UserProfile{
			ID: "1",
			ExperimentBucketMap: map[decision.UserDecisionKey]string{
				decision.NewUserDecisionKey(strValue): strValue,
			},
		}
		im.ups.Save(profile)
		i++
	}

	strValue := strconv.Itoa(2)
	expected := decision.UserProfile{
		ID: "1",
		ExperimentBucketMap: map[decision.UserDecisionKey]string{
			decision.NewUserDecisionKey(strValue): strValue,
		},
	}
	actual := im.ups.Lookup("1")
	im.Equal(expected, actual)
}

func (im *InMemoryUPSTestSuite) TestOverrideLifo() {
	im.ups = InMemoryUserProfileService{
		Order:    "lifo",
		Capacity: 10,
	}
	i := 1
	for i < 3 {
		strValue := strconv.Itoa(i)
		profile := decision.UserProfile{
			ID: "1",
			ExperimentBucketMap: map[decision.UserDecisionKey]string{
				decision.NewUserDecisionKey(strValue): strValue,
			},
		}
		im.ups.Save(profile)
		i++
	}

	strValue := strconv.Itoa(2)
	expected := decision.UserProfile{
		ID: "1",
		ExperimentBucketMap: map[decision.UserDecisionKey]string{
			decision.NewUserDecisionKey(strValue): strValue,
		},
	}
	actual := im.ups.Lookup("1")
	im.Equal(expected, actual)
}

func (im *InMemoryUPSTestSuite) TestSaveEmptyProfileFifo() {
	strValue := strconv.Itoa(1)
	profile := decision.UserProfile{
		ID: strValue,
		ExperimentBucketMap: map[decision.UserDecisionKey]string{
			decision.NewUserDecisionKey(strValue): strValue,
		},
	}
	im.ups.Save(profile)

	// Save empty profile
	profile = decision.UserProfile{
		ID:                  strValue,
		ExperimentBucketMap: map[decision.UserDecisionKey]string{},
	}
	im.ups.Save(profile)

	actual := im.ups.Lookup(strValue)
	im.Equal(profile, actual)
}

func (im *InMemoryUPSTestSuite) TestSaveEmptyProfileLifo() {
	im.ups = InMemoryUserProfileService{
		Order:    "lifo",
		Capacity: 10,
	}
	strValue := strconv.Itoa(1)
	profile := decision.UserProfile{
		ID: strValue,
		ExperimentBucketMap: map[decision.UserDecisionKey]string{
			decision.NewUserDecisionKey(strValue): strValue,
		},
	}
	im.ups.Save(profile)

	// Save empty profile
	profile = decision.UserProfile{
		ID:                  strValue,
		ExperimentBucketMap: map[decision.UserDecisionKey]string{},
	}
	im.ups.Save(profile)

	actual := im.ups.Lookup(strValue)
	im.Equal(profile, actual)
}

func (im *InMemoryUPSTestSuite) TestCapacityFifoOrEmpty() {
	// Save 10 Profiles as capacity is given as 10
	i := 1
	for i <= 10 {
		i++
		strValue := strconv.Itoa(i)
		profile := decision.UserProfile{
			ID: strValue,
			ExperimentBucketMap: map[decision.UserDecisionKey]string{
				decision.NewUserDecisionKey(strValue): strValue,
			},
		}
		im.ups.Save(profile)
	}

	// Check all 10 Profiles were saved
	i = 1
	for i <= 10 {
		i++
		strValue := strconv.Itoa(i)
		expected := decision.UserProfile{
			ID: strValue,
			ExperimentBucketMap: map[decision.UserDecisionKey]string{
				decision.NewUserDecisionKey(strValue): strValue,
			},
		}
		actual := im.ups.Lookup(strValue)
		im.Equal(expected, actual)
	}

	// Save 3 more Profiles than the capacity
	i = 11
	for i <= 13 {
		i++
		strValue := strconv.Itoa(i)
		profile := decision.UserProfile{
			ID: strValue,
			ExperimentBucketMap: map[decision.UserDecisionKey]string{
				decision.NewUserDecisionKey(strValue): strValue,
			},
		}
		im.ups.Save(profile)
	}

	// Check first 3 profiles were overwritten by newer 3 Profiles, total count still remains 10
	i = 4
	for i <= 13 {
		i++
		strValue := strconv.Itoa(i)
		expected := decision.UserProfile{
			ID: strValue,
			ExperimentBucketMap: map[decision.UserDecisionKey]string{
				decision.NewUserDecisionKey(strValue): strValue,
			},
		}
		actual := im.ups.Lookup(strValue)
		im.Equal(expected, actual)
	}
	im.Equal(10, len(im.ups.ProfilesMap))
}

func (im *InMemoryUPSTestSuite) TestCapacityLifoOrEmpty() {
	im.ups = InMemoryUserProfileService{
		Order:    "lifo",
		Capacity: 10,
	}
	// Save 10 Profiles as capacity is given as 10
	i := 1
	for i <= 10 {
		strValue := strconv.Itoa(i)
		profile := decision.UserProfile{
			ID: strValue,
			ExperimentBucketMap: map[decision.UserDecisionKey]string{
				decision.NewUserDecisionKey(strValue): strValue,
			},
		}
		im.ups.Save(profile)
		i++
	}

	// Check all 10 Profiles were saved
	i = 1
	for i <= 10 {
		strValue := strconv.Itoa(i)
		expected := decision.UserProfile{
			ID: strValue,
			ExperimentBucketMap: map[decision.UserDecisionKey]string{
				decision.NewUserDecisionKey(strValue): strValue,
			},
		}
		actual := im.ups.Lookup(strValue)
		im.Equal(expected, actual)
		i++
	}

	// Save 3 more Profiles than the capacity
	i = 11
	for i <= 13 {
		strValue := strconv.Itoa(i)
		profile := decision.UserProfile{
			ID: strValue,
			ExperimentBucketMap: map[decision.UserDecisionKey]string{
				decision.NewUserDecisionKey(strValue): strValue,
			},
		}
		im.ups.Save(profile)
		i++
	}

	// Check latest entry was always overwritten by the newer entry
	values := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 13}
	for _, v := range values {
		strValue := strconv.Itoa(v)
		expected := decision.UserProfile{
			ID: strValue,
			ExperimentBucketMap: map[decision.UserDecisionKey]string{
				decision.NewUserDecisionKey(strValue): strValue,
			},
		}
		actual := im.ups.Lookup(strValue)
		im.Equal(expected, actual)
	}
	im.Equal(10, len(im.ups.ProfilesMap))
}

func (im *InMemoryUPSTestSuite) TestZeroCapacityFifoOrEmpty() {
	im.ups = InMemoryUserProfileService{
		Capacity: 0,
	}
	// Save 200 Profiles as capacity is given as 0
	i := 1
	for i <= 200 {
		i++
		strValue := strconv.Itoa(i)
		profile := decision.UserProfile{
			ID: strValue,
			ExperimentBucketMap: map[decision.UserDecisionKey]string{
				decision.NewUserDecisionKey(strValue): strValue,
			},
		}
		im.ups.Save(profile)
	}

	// Check all 200 Profiles were saved
	i = 1
	for i <= 200 {
		i++
		strValue := strconv.Itoa(i)
		expected := decision.UserProfile{
			ID: strValue,
			ExperimentBucketMap: map[decision.UserDecisionKey]string{
				decision.NewUserDecisionKey(strValue): strValue,
			},
		}
		actual := im.ups.Lookup(strValue)
		im.Equal(expected, actual)
	}
	im.Equal(200, len(im.ups.ProfilesMap))
	im.Nil(im.ups.fifoOrderedProfiles)
}

func (im *InMemoryUPSTestSuite) TestZeroCapacityLifo() {
	im.ups = InMemoryUserProfileService{
		Order:    "lifo",
		Capacity: 0,
	}
	// Save 200 Profiles as capacity is given as 0
	i := 1
	for i <= 200 {
		i++
		strValue := strconv.Itoa(i)
		profile := decision.UserProfile{
			ID: strValue,
			ExperimentBucketMap: map[decision.UserDecisionKey]string{
				decision.NewUserDecisionKey(strValue): strValue,
			},
		}
		im.ups.Save(profile)
	}

	// Check all 200 Profiles were saved
	i = 1
	for i <= 200 {
		i++
		strValue := strconv.Itoa(i)
		expected := decision.UserProfile{
			ID: strValue,
			ExperimentBucketMap: map[decision.UserDecisionKey]string{
				decision.NewUserDecisionKey(strValue): strValue,
			},
		}
		actual := im.ups.Lookup(strValue)
		im.Equal(expected, actual)
	}
	im.Equal(200, len(im.ups.ProfilesMap))
	im.Nil(im.ups.fifoOrderedProfiles)
}

func TestInMemoryUPSTestSuite(t *testing.T) {
	suite.Run(t, new(InMemoryUPSTestSuite))
}
