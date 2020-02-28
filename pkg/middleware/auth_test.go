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
	"github.com/optimizely/agent/config"
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
	Admin     bool   `json:"admin,omitempty"`
}

func (c OptlyClaims) Valid() error {
	return nil
}

type AuthTestSuite struct {
	suite.Suite
	validAPIToken         *jwt.Token
	validAPITokenOtherSig *jwt.Token
	validAdminToken       *jwt.Token
	expiredToken          *jwt.Token
	handler               http.HandlerFunc
	signatures            []string
	authConfig            *config.ServiceAuthConfig
}

func (suite *AuthTestSuite) SetupTest() {
	suite.signatures = []string{"test", "test2"}

	claims := OptlyClaims{ExpiresAt: 12313123123213, SdkKey: "SDK_KEY", Issuer: "iss"} // exp = March 9, 2360
	suite.validAPIToken = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	suite.validAPIToken.Raw, _ = suite.validAPIToken.SignedString([]byte(suite.signatures[0]))

	claims = OptlyClaims{ExpiresAt: 12313123123213, SdkKey: "SDK_KEY", Issuer: "iss"} // exp = March 9, 2360
	suite.validAPITokenOtherSig = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	suite.validAPITokenOtherSig.Raw, _ = suite.validAPITokenOtherSig.SignedString([]byte(suite.signatures[1]))

	claims = OptlyClaims{ExpiresAt: 12313123123213, Admin: true, Issuer: "iss"} // exp = March 9, 2360
	suite.validAdminToken = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	suite.validAdminToken.Raw, _ = suite.validAdminToken.SignedString([]byte(suite.signatures[0]))

	claims = OptlyClaims{ExpiresAt: 0, SdkKey: "SDK_KEY", Issuer: "iss"}
	suite.expiredToken = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	suite.expiredToken.Raw, _ = suite.expiredToken.SignedString([]byte(suite.signatures[0]))

	suite.handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	suite.authConfig = &config.ServiceAuthConfig{
		Clients:     make([]config.OAuthClientCredentials, 0),
		HMACSecrets: suite.signatures,
		TTL:         0,
	}

}

func (suite *AuthTestSuite) TestNoAuthCheckToken() {

	auth := NewAuth(&config.ServiceAuthConfig{})
	token, err := auth.CheckToken("")
	suite.Nil(token)
	suite.NoError(err)
}

func (suite *AuthTestSuite) TestNoAuthAuthorize() {

	auth := NewAuth(&config.ServiceAuthConfig{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/some_url", nil)

	auth.AuthorizeAPI(suite.handler).ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *AuthTestSuite) TestAuthValidCheckToken() {

	auth := NewAuth(suite.authConfig)
	token, err := auth.CheckToken(suite.validAPIToken.Raw)
	suite.Equal(suite.validAPIToken.Raw, token.Raw)
	suite.NoError(err)
}

func (suite *AuthTestSuite) TestAuthInvalidCheckToken() {

	auth := NewAuth(suite.authConfig)
	token, err := auth.CheckToken("adasdsada.sfsdfs.adas")
	suite.Nil(token)
	suite.Error(err)
}

func (suite *AuthTestSuite) TestAuthAuthorizeEmptyToken() {

	auth := NewAuth(suite.authConfig)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/some_url", nil)

	auth.AuthorizeAdmin(suite.handler).ServeHTTP(rec, req)
	suite.Equal(http.StatusUnauthorized, rec.Code)
}

func (suite *AuthTestSuite) TestAuthAuthorizeAPITokenInvalidClaims() {

	auth := NewAuth(suite.authConfig)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/some_url", nil)
	req.Header.Add("Authorization", "Bearer "+suite.validAdminToken.Raw)

	auth.AuthorizeAPI(suite.handler).ServeHTTP(rec, req)
	suite.Equal(http.StatusUnauthorized, rec.Code)
}

func (suite *AuthTestSuite) TestAuthAuthorizeAdminTokenInvalidClaims() {

	auth := NewAuth(suite.authConfig)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/some_url", nil)
	req.Header.Add("Authorization", "Bearer "+suite.validAPIToken.Raw)

	auth.AuthorizeAdmin(suite.handler).ServeHTTP(rec, req)
	suite.Equal(http.StatusUnauthorized, rec.Code)
}

func (suite *AuthTestSuite) TestAuthAuthorizeAPITokenAuthorizationValidClaims() {

	auth := NewAuth(suite.authConfig)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/some_url", nil)
	req.Header.Add("Authorization", "Bearer "+suite.validAPIToken.Raw)
	req.Header.Add(OptlySDKHeader, "SDK_KEY")

	auth.AuthorizeAPI(suite.handler).ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *AuthTestSuite) TestAuthAuthorizeAPITokenAuthorizationValidClaimsOtherSig() {

	auth := NewAuth(suite.authConfig)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/some_url", nil)
	req.Header.Add("Authorization", "Bearer "+suite.validAPITokenOtherSig.Raw)
	req.Header.Add(OptlySDKHeader, "SDK_KEY")

	auth.AuthorizeAPI(suite.handler).ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *AuthTestSuite) TestAuthAuthorizeAdminTokenAuthorizationValidClaims() {

	auth := NewAuth(suite.authConfig)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/some_url", nil)
	req.Header.Add("Authorization", "Bearer "+suite.validAdminToken.Raw)
	req.Header.Add(OptlySDKHeader, "SDK_KEY")

	auth.AuthorizeAdmin(suite.handler).ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *AuthTestSuite) TestAuthAuthorizeAPITokenAuthorizationValidClaimsExpiredToken() {

	auth := NewAuth(suite.authConfig)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/some_url", nil)
	req.Header.Add("Authorization", "Bearer "+suite.expiredToken.Raw)
	req.Header.Add(OptlySDKHeader, "SDK_KEY")

	auth.AuthorizeAPI(suite.handler).ServeHTTP(rec, req)
	suite.Equal(http.StatusUnauthorized, rec.Code)
}

func (suite *AuthTestSuite) TestAuthAuthorizeAdminTokenAuthorizationValidClaimsExpiredToken() {

	auth := NewAuth(suite.authConfig)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/some_url", nil)
	req.Header.Add("Authorization", "Bearer "+suite.expiredToken.Raw)
	req.Header.Add(OptlySDKHeader, "SDK_KEY")

	auth.AuthorizeAdmin(suite.handler).ServeHTTP(rec, req)
	suite.Equal(http.StatusUnauthorized, rec.Code)
}

func TestAuth(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}
