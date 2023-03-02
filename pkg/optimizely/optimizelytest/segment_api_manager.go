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

// Package optimizelytest //
package optimizelytest

import (
	"errors"
)

// TestSegmentAPIManager implements an Optimizely segment APIManager to aid in testing
type TestSegmentAPIManager struct {
	segments  []string
	errorMode bool
	callCount int
}

// FetchQualifiedSegments returns the segments that were set by SetQualifiedSegments
func (s *TestSegmentAPIManager) FetchQualifiedSegments(apiKey, apiHost, userID string, segmentsToCheck []string) ([]string, error) {
	s.callCount += 1
	if s.errorMode {
		return nil, errors.New("failed to fetch qualified segments")
	}
	return s.segments, nil
}

// SetQualifiedSegments determines the segments that will be returned by FetchQualifiedSegments
func (s *TestSegmentAPIManager) SetQualifiedSegments(segments []string) {
	s.segments = segments
}

// SetErrorMode determines if FetchQualifiedSegments returns an error or not.
func (s *TestSegmentAPIManager) SetErrorMode(errorMode bool) {
	s.errorMode = errorMode
}

// GetCallCount returns the number of times FetchQualifiedSegments has been called.
func (s *TestSegmentAPIManager) GetCallCount() int {
	return s.callCount
}
