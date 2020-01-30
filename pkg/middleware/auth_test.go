/****************************************************************************
 * Copyright 2020, Optimizely, Inc. and contributors                        *
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

// Package middleware //
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/suite"
)

type OptlyClaims struct {
	ExpiresAt int64  `json:"exp,omitempty"`
	Issuer    string `json:"iss,omitempty"`
	SdkKey    string `json:"sdk_key,omitempty"`
}

func (c OptlyClaims) Valid() error {
	return nil
}

type AuthTestSuite struct {
	suite.Suite
	validToken   *jwt.Token
	expiredToken *jwt.Token
	handler      http.HandlerFunc
	signature    string
}

func (suite *AuthTestSuite) SetupTest() {
	suite.signature = "test"
	claims := OptlyClaims{ExpiresAt: 12313123123213, SdkKey: "SDK_KEY", Issuer: "iss"} // exp = March 9, 2360
	suite.validToken = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	suite.validToken.Raw, _ = suite.validToken.SignedString([]byte(suite.signature))

	claims = OptlyClaims{ExpiresAt: 0, SdkKey: "SDK_KEY", Issuer: "iss"}
	suite.expiredToken = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	suite.expiredToken.Raw, _ = suite.expiredToken.SignedString([]byte(suite.signature))

	suite.handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

}

func (suite *AuthTestSuite) TestNoAuthCheckToken() {

	auth := NewAuth(NoAuth{}, map[string]struct{}{})
	token, err := auth.CheckToken("")
	suite.Equal(nil, token)
	suite.NoError(err)
}

func (suite *AuthTestSuite) TestNoAuthAuthorize() {

	auth := NewAuth(NoAuth{}, map[string]struct{}{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/some_url", nil)

	auth.Authorize(suite.handler).ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *AuthTestSuite) TestAuthValidCheckToken() {

	auth := NewAuth(NewJWTVerifier(suite.signature), map[string]struct{}{})
	token, err := auth.CheckToken(suite.validToken.Raw)
	suite.Equal(suite.validToken.Raw, token.Raw)
	suite.NoError(err)
}

func (suite *AuthTestSuite) TestAuthInvalidCheckToken() {

	auth := NewAuth(NewJWTVerifier(suite.signature), map[string]struct{}{})
	token, err := auth.CheckToken("adasdsada.sfsdfs.adas")
	suite.Nil(token)
	suite.Error(err)
}

func (suite *AuthTestSuite) TestAuthAuthorizeEmptyToken() {

	auth := NewAuth(NewJWTVerifier(suite.signature), map[string]struct{}{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/some_url", nil)

	auth.Authorize(suite.handler).ServeHTTP(rec, req)
	suite.Equal(http.StatusUnauthorized, rec.Code)
}

func (suite *AuthTestSuite) TestAuthAuthorizeToken() {

	auth := NewAuth(NewJWTVerifier(suite.signature), map[string]struct{}{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/some_url", nil)
	req.Header.Add("Auth", suite.validToken.Raw)
	auth.Authorize(suite.handler).ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *AuthTestSuite) TestAuthAuthorizeTokenAuthorization() {

	auth := NewAuth(NewJWTVerifier(suite.signature), map[string]struct{}{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/some_url", nil)
	req.Header.Add("Authorization", "Bearer "+suite.validToken.Raw)

	auth.Authorize(suite.handler).ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *AuthTestSuite) TestAuthAuthorizeTokenInvalidClaims() {

	auth := NewAuth(NewJWTVerifier(suite.signature), map[string]struct{}{"sdk_key": {}})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/some_url", nil)
	req.Header.Add("Authorization", "Bearer "+suite.validToken.Raw)

	auth.Authorize(suite.handler).ServeHTTP(rec, req)
	suite.Equal(http.StatusUnauthorized, rec.Code)
}

func (suite *AuthTestSuite) TestAuthAuthorizeTokenAuthorizationValidClaims() {

	auth := NewAuth(NewJWTVerifier(suite.signature), map[string]struct{}{"sdk_key": {}, "exp": {}})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/some_url", nil)
	req.Header.Add("Authorization", "Bearer "+suite.validToken.Raw)
	req.Header.Add(OptlySDKHeader, "SDK_KEY")

	auth.Authorize(suite.handler).ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *AuthTestSuite) TestAuthAuthorizeTokenAuthorizationValidClaimsExpiredToken() {

	auth := NewAuth(NewJWTVerifier(suite.signature), map[string]struct{}{"exp": {}})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/some_url", nil)
	req.Header.Add("Authorization", "Bearer "+suite.expiredToken.Raw)
	req.Header.Add(OptlySDKHeader, "SDK_KEY")

	auth.Authorize(suite.handler).ServeHTTP(rec, req)
	suite.Equal(http.StatusUnauthorized, rec.Code)
}

func TestAuth(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}
