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
	"os"
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
	validAPIToken   *jwt.Token
	validAdminToken *jwt.Token
	expiredToken    *jwt.Token
	handler         http.HandlerFunc
	signature       string
	authConfig      *config.ServiceAuthConfig
}

func (suite *AuthTestSuite) SetupTest() {
	suite.signature = "test"
	claims := OptlyClaims{ExpiresAt: 12313123123213, SdkKey: "SDK_KEY", Issuer: "iss"} // exp = March 9, 2360
	suite.validAPIToken = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	suite.validAPIToken.Raw, _ = suite.validAPIToken.SignedString([]byte(suite.signature))

	claims = OptlyClaims{ExpiresAt: 12313123123213, Admin: true, Issuer: "iss"} // exp = March 9, 2360
	suite.validAdminToken = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	suite.validAdminToken.Raw, _ = suite.validAdminToken.SignedString([]byte(suite.signature))

	claims = OptlyClaims{ExpiresAt: 0, SdkKey: "SDK_KEY", Issuer: "iss"}
	suite.expiredToken = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	suite.expiredToken.Raw, _ = suite.expiredToken.SignedString([]byte(suite.signature))

	suite.handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	suite.authConfig = &config.ServiceAuthConfig{
		Clients:    make([]config.OAuthClientCredentials, 0),
		HMACSecret: suite.signature,
		TTL:        0,
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

func (suite *AuthTestSuite) TestAuthValidCheckTokenFromJwks() {

	const tk = `eyJhbGciOiJSUzI1NiIsImtpZCI6Il9ZdXhXVHgyZHAyRVNVb2s3MmUzcjNLb0R6OWZueFdJM29DQndOYnkyX0UiLCJ0eXAiOiJKV1QifQ.eyJhY2NvdW50X2lkIjo0Njg1MjgwNDQ0LCJhdWQiOiJTUlZDIiwiZXhwIjoxNTgyMzI1NjE5LCJpYXQiOjE1ODIzMjUwMTksImlzcyI6IlRPS0VOX1NFUlZJQ0UiLCJqdGkiOiI2OWVlN2M2NS1jNWU1LTQwNzYtYjI2Zi0yOGYzY2JlZjQwZjUiLCJuYmYiOjE1ODIzMjUwMTksInByb2plY3RfaWQiOjkyNjQzNjc2OTAsInNjb3BlcyI6ImF0dHJpYnV0ZXMubW9kaWZ5IGF0dHJpYnV0ZXMucmVhZCBhdWRpZW5jZXMucmVhZCBjaGFuZ2VfaGlzdG9yeS5yZWFkIGNvbGxhYm9yYXRvcnMubW9kaWZ5IGNvbGxhYm9yYXRvcnMucmVhZCBkY3AubW9kaWZ5IGRjcC5yZWFkIGV2ZW50cy5yZWFkIGV4cGVyaW1lbnRzLm1vZGlmeSBleHBlcmltZW50cy5yZWFkIGV4dGVuc2lvbnMubW9kaWZ5IGV4dGVuc2lvbnMucmVhZCBwYWdlcy5yZWFkIHByb2plY3RzLnJlYWQgcmVjb21tZW5kZXJzLm1vZGlmeSByZWNvbW1lbmRlcnMucmVhZCByZXN1bHRzLnJlYWQgc2FyLm1vZGlmeSBzYXIucmVhZCB1c2VyLnJlYWQiLCJzdWIiOiJ1cm46dXNlcjplMDk4ZjYwMGMwNzkxMWU1YjQ0NmU1MGRmMzdhOWEyMiIsInVzZXJfaWQiOiJlMDk4ZjYwMGMwNzkxMWU1YjQ0NmU1MGRmMzdhOWEyMiJ9.D9KVOiyvMP8ctJHIJAjt1ddj4Dol1c7vmPc0ZJg9A7t-yOv3WDjlxYMeTOPwPvN3iTHxIb-MFGIQyDpv63v13s00G0P4CFJHdXYBYTQETHCH1kFfjU5hK1lUAlqel3v25-uE-LgOnpnDsJK_LBmPwGJxh1_S5lyY6fBpQo9guMgmFIoN-GXGHzSWMD93oyD5CoiXWbxvLMIGMOrafl3YzqnEPK4WgmujnSR2vnj5lSLuJF_5-EICSXwuK2JVOq0xjGwa2trhw6xeVzN7JcKMb_baRq2tKxiiOjTnC-jPtkR22G8CWFcWUtOkkl-9XM9PXop2tHyLDWXxk73RChpAHg`
	dir, _ := os.Getwd()
	authConfig := &config.ServiceAuthConfig{
		Clients:    make([]config.OAuthClientCredentials, 0),
		HMACSecret: suite.signature,
		TTL:        0,
		JwksURL:    "file://" + dir + "/testdata/jwks_url.txt",
	}

	auth := JWTVerifierURL{jwksURL: authConfig.JwksURL, parser: &jwt.Parser{SkipClaimsValidation: true}}
	token, err := auth.CheckToken(tk)
	suite.Equal(tk, token.Raw)
	suite.NoError(err)
}

func (suite *AuthTestSuite) TestAuthInvalidCheckTokenFromJwksURL() {

	const tk = `eyJhbGciOiJSUzI1NiIsImtpZCI6Il9ZdXhXVHgyZHAyRVNVb2s3MmUzcjNLb0R6OWZueFdJM29DQndOYnkyX0UiLCJ0eXAiOiJKV1QifQ.eyJhY2NvdW50X2lkIjo0Njg1MjgwNDQ0LCJhdWQiOiJTUlZDIiwiZXhwIjoxNTgyMzI1NjE5LCJpYXQiOjE1ODIzMjUwMTksImlzcyI6IlRPS0VOX1NFUlZJQ0UiLCJqdGkiOiI2OWVlN2M2NS1jNWU1LTQwNzYtYjI2Zi0yOGYzY2JlZjQwZjUiLCJuYmYiOjE1ODIzMjUwMTksInByb2plY3RfaWQiOjkyNjQzNjc2OTAsInNjb3BlcyI6ImF0dHJpYnV0ZXMubW9kaWZ5IGF0dHJpYnV0ZXMucmVhZCBhdWRpZW5jZXMucmVhZCBjaGFuZ2VfaGlzdG9yeS5yZWFkIGNvbGxhYm9yYXRvcnMubW9kaWZ5IGNvbGxhYm9yYXRvcnMucmVhZCBkY3AubW9kaWZ5IGRjcC5yZWFkIGV2ZW50cy5yZWFkIGV4cGVyaW1lbnRzLm1vZGlmeSBleHBlcmltZW50cy5yZWFkIGV4dGVuc2lvbnMubW9kaWZ5IGV4dGVuc2lvbnMucmVhZCBwYWdlcy5yZWFkIHByb2plY3RzLnJlYWQgcmVjb21tZW5kZXJzLm1vZGlmeSByZWNvbW1lbmRlcnMucmVhZCByZXN1bHRzLnJlYWQgc2FyLm1vZGlmeSBzYXIucmVhZCB1c2VyLnJlYWQiLCJzdWIiOiJ1cm46dXNlcjplMDk4ZjYwMGMwNzkxMWU1YjQ0NmU1MGRmMzdhOWEyMiIsInVzZXJfaWQiOiJlMDk4ZjYwMGMwNzkxMWU1YjQ0NmU1MGRmMzdhOWEyMiJ9.D9KVOiyvMP8ctJHIJAjt1ddj4Dol1c7vmPc0ZJg9A7t-yOv3WDjlxYMeTOPwPvN3iTHxIb-MFGIQyDpv63v13s00G0P4CFJHdXYBYTQETHCH1kFfjU5hK1lUAlqel3v25-uE-LgOnpnDsJK_LBmPwGJxh1_S5lyY6fBpQo9guMgmFIoN-GXGHzSWMD93oyD5CoiXWbxvLMIGMOrafl3YzqnEPK4WgmujnSR2vnj5lSLuJF_5-EICSXwuK2JVOq0xjGwa2trhw6xeVzN7JcKMb_baRq2tKxiiOjTnC-jPtkR22G8CWFcWUtOkkl-9XM9PXop2tHyLDWXxk73RChpAHg`

	authConfig := &config.ServiceAuthConfig{
		Clients:    make([]config.OAuthClientCredentials, 0),
		HMACSecret: suite.signature,
		TTL:        0,
		JwksURL:    "fake_url",
	}

	auth := NewAuth(authConfig)
	token, err := auth.CheckToken(tk)
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

func (suite *AuthTestSuite) TestAuthAuthorizeInvalidJwksURL() {

	authConfig := &config.ServiceAuthConfig{
		Clients:    make([]config.OAuthClientCredentials, 0),
		HMACSecret: suite.signature,
		TTL:        0,
		JwksURL:    "fake_url",
	}

	auth := NewAuth(authConfig)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/some_url", nil)
	req.Header.Add("Authorization", "Bearer "+suite.validAdminToken.Raw)
	req.Header.Add(OptlySDKHeader, "SDK_KEY")

	auth.AuthorizeAdmin(suite.handler).ServeHTTP(rec, req)
	suite.Equal(http.StatusUnauthorized, rec.Code)
}

func (suite *AuthTestSuite) TestAuthAuthorizeMockValidJwksURL() {
	const token = `eyJhbGciOiJSUzI1NiIsImtpZCI6Il9ZdXhXVHgyZHAyRVNVb2s3MmUzcjNLb0R6OWZueFdJM29DQndOYnkyX0UiLCJ0eXAiOiJKV1QifQ.eyJhY2NvdW50X2lkIjo0Njg1MjgwNDQ0LCJhdWQiOiJTUlZDIiwiZXhwIjoxNTgyMzI1NjE5LCJpYXQiOjE1ODIzMjUwMTksImlzcyI6IlRPS0VOX1NFUlZJQ0UiLCJqdGkiOiI2OWVlN2M2NS1jNWU1LTQwNzYtYjI2Zi0yOGYzY2JlZjQwZjUiLCJuYmYiOjE1ODIzMjUwMTksInByb2plY3RfaWQiOjkyNjQzNjc2OTAsInNjb3BlcyI6ImF0dHJpYnV0ZXMubW9kaWZ5IGF0dHJpYnV0ZXMucmVhZCBhdWRpZW5jZXMucmVhZCBjaGFuZ2VfaGlzdG9yeS5yZWFkIGNvbGxhYm9yYXRvcnMubW9kaWZ5IGNvbGxhYm9yYXRvcnMucmVhZCBkY3AubW9kaWZ5IGRjcC5yZWFkIGV2ZW50cy5yZWFkIGV4cGVyaW1lbnRzLm1vZGlmeSBleHBlcmltZW50cy5yZWFkIGV4dGVuc2lvbnMubW9kaWZ5IGV4dGVuc2lvbnMucmVhZCBwYWdlcy5yZWFkIHByb2plY3RzLnJlYWQgcmVjb21tZW5kZXJzLm1vZGlmeSByZWNvbW1lbmRlcnMucmVhZCByZXN1bHRzLnJlYWQgc2FyLm1vZGlmeSBzYXIucmVhZCB1c2VyLnJlYWQiLCJzdWIiOiJ1cm46dXNlcjplMDk4ZjYwMGMwNzkxMWU1YjQ0NmU1MGRmMzdhOWEyMiIsInVzZXJfaWQiOiJlMDk4ZjYwMGMwNzkxMWU1YjQ0NmU1MGRmMzdhOWEyMiJ9.D9KVOiyvMP8ctJHIJAjt1ddj4Dol1c7vmPc0ZJg9A7t-yOv3WDjlxYMeTOPwPvN3iTHxIb-MFGIQyDpv63v13s00G0P4CFJHdXYBYTQETHCH1kFfjU5hK1lUAlqel3v25-uE-LgOnpnDsJK_LBmPwGJxh1_S5lyY6fBpQo9guMgmFIoN-GXGHzSWMD93oyD5CoiXWbxvLMIGMOrafl3YzqnEPK4WgmujnSR2vnj5lSLuJF_5-EICSXwuK2JVOq0xjGwa2trhw6xeVzN7JcKMb_baRq2tKxiiOjTnC-jPtkR22G8CWFcWUtOkkl-9XM9PXop2tHyLDWXxk73RChpAHg`

	authConfig := &config.ServiceAuthConfig{
		Clients:    make([]config.OAuthClientCredentials, 0),
		HMACSecret: suite.signature,
		TTL:        0,
		JwksURL:    "fake_url",
	}

	auth := NewAuth(authConfig)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/some_url", nil)
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add(OptlySDKHeader, "SDK_KEY")

	auth.AuthorizeAdmin(suite.handler).ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)
}

func TestAuth(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}
