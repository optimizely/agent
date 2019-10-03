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

// Package optimizely //
package optimizely

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ContextTestSuite struct {
	suite.Suite
	optlyContext *OptlyContext
}

func (suite *ContextTestSuite) SetupTest() {
	suite.optlyContext = NewContext("userId", map[string]interface{}{"key": "val"})
}

func (suite *ContextTestSuite) TestUserContext() {
	suite.Equal("userId", suite.optlyContext.UserContext.ID)
	suite.Equal("val", suite.optlyContext.UserContext.Attributes["key"])
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestContextTestSuite(t *testing.T) {
	suite.Run(t, new(ContextTestSuite))
}
