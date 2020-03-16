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
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/optimizely/agent/config"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type OptlyClaims struct {
	ExpiresAt int64    `json:"exp,omitempty"`
	Issuer    string   `json:"iss,omitempty"`
	SdkKeys   []string `json:"sdk_keys,omitempty"`
	Admin     bool     `json:"admin,omitempty"`
}

func (c OptlyClaims) Valid() error {
	return nil
}

type AuthTestSuite struct {
	suite.Suite

	server *httptest.Server

	validAPIToken            *jwt.Token
	validAPITokenOtherSig    *jwt.Token
	validAPITokenMultiSdkKey *jwt.Token
	validAdminToken          *jwt.Token
	expiredToken             *jwt.Token
	handler                  http.HandlerFunc
	signatures               []string
	authConfig               *config.ServiceAuthConfig
}

func (suite *AuthTestSuite) SetupTest() {
	suite.signatures = []string{"R8W3PRpnjp6/WmhyeCBZdscrQbMpqf8WIDxx910SlJk=", "LQR5YqSDln9ALMg7PsxC0/69ktrCJEPPUG4gwzZHAww="}

	var decodedSigs [][]byte
	for _, sig := range suite.signatures {
		decoded, _ := base64.StdEncoding.DecodeString(sig)
		decodedSigs = append(decodedSigs, decoded)
	}

	claims := OptlyClaims{ExpiresAt: 12313123123213, SdkKeys: []string{"SDK_KEY"}, Issuer: "iss"} // exp = March 9, 2360
	suite.validAPIToken = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	suite.validAPIToken.Raw, _ = suite.validAPIToken.SignedString(decodedSigs[0])

	claims = OptlyClaims{ExpiresAt: 12313123123213, SdkKeys: []string{"SDK_KEY"}, Issuer: "iss"} // exp = March 9, 2360
	suite.validAPITokenOtherSig = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	suite.validAPITokenOtherSig.Raw, _ = suite.validAPITokenOtherSig.SignedString(decodedSigs[1])

	claims = OptlyClaims{ExpiresAt: 12313123123213, Admin: true, Issuer: "iss"} // exp = March 9, 2360
	suite.validAdminToken = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	suite.validAdminToken.Raw, _ = suite.validAdminToken.SignedString(decodedSigs[0])

	claims = OptlyClaims{ExpiresAt: 0, SdkKeys: []string{"SDK_KEY"}, Issuer: "iss"}
	suite.expiredToken = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	suite.expiredToken.Raw, _ = suite.expiredToken.SignedString(decodedSigs[0])

	claims = OptlyClaims{ExpiresAt: 12313123123213, SdkKeys: []string{"SDK_KEY_1", "SDK_KEY_2"}, Issuer: "iss"} // exp = March 9, 2360
	suite.validAPITokenMultiSdkKey = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	suite.validAPITokenMultiSdkKey.Raw, _ = suite.validAPITokenMultiSdkKey.SignedString(decodedSigs[0])

	suite.handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	suite.authConfig = &config.ServiceAuthConfig{
		Clients:     make([]config.OAuthClientCredentials, 0),
		HMACSecrets: suite.signatures,
		TTL:         0,
	}

	suite.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.String() == "/good" {
			fmt.Fprintln(w, `{"keys":[{"alg":"RS256","e":"AQAB","kid":"_YuxWTx2dp2ESUok72e3r3KoDz9fnxWI3oCBwNby2_E","kty":"RSA","n":"2ZNUw2VOO30mR15JcT5Lz85GznV2p3K0DtXRJiOhGOD0YnCkNZL3cPHR_r7_eVVJMokz4yGIW8hwSJN0GrzmihzULDTpFlAmSkSissSMIYANZOdHOPm5iCYCCeX_5ceCtDS85Z2gh0dN7vX7GkoYxJs-eLc0W8EVzA5V8S9c42ARGenH99nX8CiwiINEoZyLvv-Le2RX5zetVWqVD6EfP-mjzku-h5Nxx4PLk8tdiSpV-DllVGoYt5_P9_FgyTsZ1-62e2GJmNy0odZEUsTAxWnF_c1InEQZggI-vtCPNNVF1qgjArc86mGBc6z26EmRU91TavehP6n_oszhif83QQ","use":"sig"}]}`)
		}
		if r.URL.String() == "/bad" {
			fmt.Fprintln(w, `{"keys":[{"alg":"RS256","e":"AQAB","kid":"bad_id","kty":"RSA","n":"2ZNUw2VOO30mR15JcT5Lz85GznV2p3K0DtXRJiOhGOD0YnCkNZL3cPHR_r7_eVVJMokz4yGIW8hwSJN0GrzmihzULDTpFlAmSkSissSMIYANZOdHOPm5iCYCCeX_5ceCtDS85Z2gh0dN7vX7GkoYxJs-eLc0W8EVzA5V8S9c42ARGenH99nX8CiwiINEoZyLvv-Le2RX5zetVWqVD6EfP-mjzku-h5Nxx4PLk8tdiSpV-DllVGoYt5_P9_FgyTsZ1-62e2GJmNy0odZEUsTAxWnF_c1InEQZggI-vtCPNNVF1qgjArc86mGBc6z26EmRU91TavehP6n_oszhif83QQ","use":"sig"}]}`)
		}
	}))

}

func (suite *AuthTestSuite) TearDownTest() {
	suite.server.Close()

}

func (suite *AuthTestSuite) TestNewAuthNoAuth() {
	authConfig := &config.ServiceAuthConfig{}
	auth := NewAuth(authConfig)

	if _, ok := auth.Verifier.(NoAuth); !ok {
		suite.Fail("expected NoAuth type")
	}
}

func (suite *AuthTestSuite) TestNewAuthJWTVerifier() {
	authConfig := &config.ServiceAuthConfig{
		Clients:     make([]config.OAuthClientCredentials, 0),
		HMACSecrets: suite.signatures,
		TTL:         0,
	}
	auth := NewAuth(authConfig)

	if _, ok := auth.Verifier.(*JWTVerifier); !ok {
		suite.Fail("expected JWTVerifier type")
	}
}

func (suite *AuthTestSuite) TestNewAuthJWTVerifierURL() {
	authConfig := &config.ServiceAuthConfig{
		Clients:            make([]config.OAuthClientCredentials, 0),
		HMACSecrets:        suite.signatures,
		TTL:                0,
		JwksURL:            suite.server.URL + "/good",
		JwksUpdateInterval: time.Second,
	}
	auth := NewAuth(authConfig)

	if _, ok := auth.Verifier.(*JWTVerifierURL); !ok {
		suite.Fail("expected JWTVerifierURL type")
	}
}

func (suite *AuthTestSuite) TestNewAuthBadAuthNoInterval() {
	authConfig := &config.ServiceAuthConfig{
		Clients:     make([]config.OAuthClientCredentials, 0),
		HMACSecrets: suite.signatures,
		TTL:         0,
		JwksURL:     suite.server.URL + "/good",
	}

	auth := NewAuth(authConfig)
	suite.Nil(auth)
}

func (suite *AuthTestSuite) TestNewAuthBadAuthBadURL() {
	authConfig := &config.ServiceAuthConfig{
		Clients:            make([]config.OAuthClientCredentials, 0),
		HMACSecrets:        suite.signatures,
		TTL:                0,
		JwksURL:            "fake_url",
		JwksUpdateInterval: time.Second,
	}

	auth := NewAuth(authConfig)
	suite.Nil(auth)
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

func (suite *AuthTestSuite) TestAuthValidCheckTokenFromValidJwks() {

	const tk = `eyJhbGciOiJSUzI1NiIsImtpZCI6Il9ZdXhXVHgyZHAyRVNVb2s3MmUzcjNLb0R6OWZueFdJM29DQndOYnkyX0UiLCJ0eXAiOiJKV1QifQ.eyJhY2NvdW50X2lkIjo0Njg1MjgwNDQ0LCJhdWQiOiJTUlZDIiwiZXhwIjoxNTgyMzI1NjE5LCJpYXQiOjE1ODIzMjUwMTksImlzcyI6IlRPS0VOX1NFUlZJQ0UiLCJqdGkiOiI2OWVlN2M2NS1jNWU1LTQwNzYtYjI2Zi0yOGYzY2JlZjQwZjUiLCJuYmYiOjE1ODIzMjUwMTksInByb2plY3RfaWQiOjkyNjQzNjc2OTAsInNjb3BlcyI6ImF0dHJpYnV0ZXMubW9kaWZ5IGF0dHJpYnV0ZXMucmVhZCBhdWRpZW5jZXMucmVhZCBjaGFuZ2VfaGlzdG9yeS5yZWFkIGNvbGxhYm9yYXRvcnMubW9kaWZ5IGNvbGxhYm9yYXRvcnMucmVhZCBkY3AubW9kaWZ5IGRjcC5yZWFkIGV2ZW50cy5yZWFkIGV4cGVyaW1lbnRzLm1vZGlmeSBleHBlcmltZW50cy5yZWFkIGV4dGVuc2lvbnMubW9kaWZ5IGV4dGVuc2lvbnMucmVhZCBwYWdlcy5yZWFkIHByb2plY3RzLnJlYWQgcmVjb21tZW5kZXJzLm1vZGlmeSByZWNvbW1lbmRlcnMucmVhZCByZXN1bHRzLnJlYWQgc2FyLm1vZGlmeSBzYXIucmVhZCB1c2VyLnJlYWQiLCJzdWIiOiJ1cm46dXNlcjplMDk4ZjYwMGMwNzkxMWU1YjQ0NmU1MGRmMzdhOWEyMiIsInVzZXJfaWQiOiJlMDk4ZjYwMGMwNzkxMWU1YjQ0NmU1MGRmMzdhOWEyMiJ9.D9KVOiyvMP8ctJHIJAjt1ddj4Dol1c7vmPc0ZJg9A7t-yOv3WDjlxYMeTOPwPvN3iTHxIb-MFGIQyDpv63v13s00G0P4CFJHdXYBYTQETHCH1kFfjU5hK1lUAlqel3v25-uE-LgOnpnDsJK_LBmPwGJxh1_S5lyY6fBpQo9guMgmFIoN-GXGHzSWMD93oyD5CoiXWbxvLMIGMOrafl3YzqnEPK4WgmujnSR2vnj5lSLuJF_5-EICSXwuK2JVOq0xjGwa2trhw6xeVzN7JcKMb_baRq2tKxiiOjTnC-jPtkR22G8CWFcWUtOkkl-9XM9PXop2tHyLDWXxk73RChpAHg`

	authConfig := &config.ServiceAuthConfig{
		Clients:     make([]config.OAuthClientCredentials, 0),
		HMACSecrets: suite.signatures,
		TTL:         0,
		JwksURL:     suite.server.URL + "/good",
	}

	auth := JWTVerifierURL{jwksURL: authConfig.JwksURL, parser: &jwt.Parser{SkipClaimsValidation: true}}

	auth.updateKeySet()
	token, err := auth.CheckToken(tk)
	suite.Equal(tk, token.Raw)
	suite.NoError(err)
}

func (suite *AuthTestSuite) TestAuthValidCheckTokenFromInvalidJwksURL() {

	const tk = `eyJhbGciOiJSUzI1NiIsImtpZCI6Il9ZdXhXVHgyZHAyRVNVb2s3MmUzcjNLb0R6OWZueFdJM29DQndOYnkyX0UiLCJ0eXAiOiJKV1QifQ.eyJhY2NvdW50X2lkIjo0Njg1MjgwNDQ0LCJhdWQiOiJTUlZDIiwiZXhwIjoxNTgyMzI1NjE5LCJpYXQiOjE1ODIzMjUwMTksImlzcyI6IlRPS0VOX1NFUlZJQ0UiLCJqdGkiOiI2OWVlN2M2NS1jNWU1LTQwNzYtYjI2Zi0yOGYzY2JlZjQwZjUiLCJuYmYiOjE1ODIzMjUwMTksInByb2plY3RfaWQiOjkyNjQzNjc2OTAsInNjb3BlcyI6ImF0dHJpYnV0ZXMubW9kaWZ5IGF0dHJpYnV0ZXMucmVhZCBhdWRpZW5jZXMucmVhZCBjaGFuZ2VfaGlzdG9yeS5yZWFkIGNvbGxhYm9yYXRvcnMubW9kaWZ5IGNvbGxhYm9yYXRvcnMucmVhZCBkY3AubW9kaWZ5IGRjcC5yZWFkIGV2ZW50cy5yZWFkIGV4cGVyaW1lbnRzLm1vZGlmeSBleHBlcmltZW50cy5yZWFkIGV4dGVuc2lvbnMubW9kaWZ5IGV4dGVuc2lvbnMucmVhZCBwYWdlcy5yZWFkIHByb2plY3RzLnJlYWQgcmVjb21tZW5kZXJzLm1vZGlmeSByZWNvbW1lbmRlcnMucmVhZCByZXN1bHRzLnJlYWQgc2FyLm1vZGlmeSBzYXIucmVhZCB1c2VyLnJlYWQiLCJzdWIiOiJ1cm46dXNlcjplMDk4ZjYwMGMwNzkxMWU1YjQ0NmU1MGRmMzdhOWEyMiIsInVzZXJfaWQiOiJlMDk4ZjYwMGMwNzkxMWU1YjQ0NmU1MGRmMzdhOWEyMiJ9.D9KVOiyvMP8ctJHIJAjt1ddj4Dol1c7vmPc0ZJg9A7t-yOv3WDjlxYMeTOPwPvN3iTHxIb-MFGIQyDpv63v13s00G0P4CFJHdXYBYTQETHCH1kFfjU5hK1lUAlqel3v25-uE-LgOnpnDsJK_LBmPwGJxh1_S5lyY6fBpQo9guMgmFIoN-GXGHzSWMD93oyD5CoiXWbxvLMIGMOrafl3YzqnEPK4WgmujnSR2vnj5lSLuJF_5-EICSXwuK2JVOq0xjGwa2trhw6xeVzN7JcKMb_baRq2tKxiiOjTnC-jPtkR22G8CWFcWUtOkkl-9XM9PXop2tHyLDWXxk73RChpAHg`

	authConfig := &config.ServiceAuthConfig{
		Clients:     make([]config.OAuthClientCredentials, 0),
		HMACSecrets: suite.signatures,
		TTL:         0,
		JwksURL:     "fake_url",
	}

	auth := NewAuth(authConfig)
	suite.Nil(auth)

}

func (suite *AuthTestSuite) TestAuthInvalidCheckTokenFromValidJwksURL() {

	const tk = `eyJhbGciOiJSUzI1NiIsImtpZCI6Il9ZdXhXVHgyZHAyRVNVb2s3MmUzcjNLb0R6OWZueFdJM29DQndOYnkyX0UiLCJ0eXAiOiJKV1QifQ.eyJhY2NvdW50X2lkIjo0Njg1MjgwNDQ0LCJhdWQiOiJTUlZDIiwiZXhwIjoxNTgyMzI1NjE5LCJpYXQiOjE1ODIzMjUwMTksImlzcyI6IlRPS0VOX1NFUlZJQ0UiLCJqdGkiOiI2OWVlN2M2NS1jNWU1LTQwNzYtYjI2Zi0yOGYzY2JlZjQwZjUiLCJuYmYiOjE1ODIzMjUwMTksInByb2plY3RfaWQiOjkyNjQzNjc2OTAsInNjb3BlcyI6ImF0dHJpYnV0ZXMubW9kaWZ5IGF0dHJpYnV0ZXMucmVhZCBhdWRpZW5jZXMucmVhZCBjaGFuZ2VfaGlzdG9yeS5yZWFkIGNvbGxhYm9yYXRvcnMubW9kaWZ5IGNvbGxhYm9yYXRvcnMucmVhZCBkY3AubW9kaWZ5IGRjcC5yZWFkIGV2ZW50cy5yZWFkIGV4cGVyaW1lbnRzLm1vZGlmeSBleHBlcmltZW50cy5yZWFkIGV4dGVuc2lvbnMubW9kaWZ5IGV4dGVuc2lvbnMucmVhZCBwYWdlcy5yZWFkIHByb2plY3RzLnJlYWQgcmVjb21tZW5kZXJzLm1vZGlmeSByZWNvbW1lbmRlcnMucmVhZCByZXN1bHRzLnJlYWQgc2FyLm1vZGlmeSBzYXIucmVhZCB1c2VyLnJlYWQiLCJzdWIiOiJ1cm46dXNlcjplMDk4ZjYwMGMwNzkxMWU1YjQ0NmU1MGRmMzdhOWEyMiIsInVzZXJfaWQiOiJlMDk4ZjYwMGMwNzkxMWU1YjQ0NmU1MGRmMzdhOWEyMiJ9.D9KVOiyvMP8ctJHIJAjt1ddj4Dol1c7vmPc0ZJg9A7t-yOv3WDjlxYMeTOPwPvN3iTHxIb-MFGIQyDpv63v13s00G0P4CFJHdXYBYTQETHCH1kFfjU5hK1lUAlqel3v25-uE-LgOnpnDsJK_LBmPwGJxh1_S5lyY6fBpQo9guMgmFIoN-GXGHzSWMD93oyD5CoiXWbxvLMIGMOrafl3YzqnEPK4WgmujnSR2vnj5lSLuJF_5-EICSXwuK2JVOq0xjGwa2trhw6xeVzN7JcKMb_baRq2tKxiiOjTnC-jPtkR22G8CWFcWUtOkkl-9XM9PXop2tHyLDWXxk73RChpAHg_invalid`

	authConfig := &config.ServiceAuthConfig{
		Clients:     make([]config.OAuthClientCredentials, 0),
		HMACSecrets: suite.signatures,
		TTL:         0,
		JwksURL:     suite.server.URL + "/good",
	}

	auth := JWTVerifierURL{jwksURL: authConfig.JwksURL, parser: &jwt.Parser{SkipClaimsValidation: true}}
	auth.updateKeySet()
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

func (suite *AuthTestSuite) TestAuthAuthorizeAPITokenAuthorizationValidClaimsOtherSig() {

	auth := NewAuth(suite.authConfig)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/some_url", nil)
	req.Header.Add("Authorization", "Bearer "+suite.validAPITokenOtherSig.Raw)
	req.Header.Add(OptlySDKHeader, "SDK_KEY")

	auth.AuthorizeAPI(suite.handler).ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)
}

func (suite *AuthTestSuite) TestAuthAuthorizeAPITokenInvalidHeaderSDKKey() {
	auth := NewAuth(suite.authConfig)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/some_url", nil)
	req.Header.Add("Authorization", "Bearer "+suite.validAPIToken.Raw)
	req.Header.Add(OptlySDKHeader, "OTHER_SDK_KEY")

	auth.AuthorizeAPI(suite.handler).ServeHTTP(rec, req)
	suite.Equal(http.StatusUnauthorized, rec.Code)
}

func (suite *AuthTestSuite) TestAuthAuthorizeAPITokenValidClaimsMultipleSDKKeys() {
	auth := NewAuth(suite.authConfig)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/some_url", nil)
	req.Header.Add("Authorization", "Bearer "+suite.validAPITokenMultiSdkKey.Raw)
	req.Header.Add(OptlySDKHeader, "SDK_KEY_1")
	auth.AuthorizeAPI(suite.handler).ServeHTTP(rec, req)
	suite.Equal(http.StatusOK, rec.Code)

	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/some_url", nil)
	req.Header.Add("Authorization", "Bearer "+suite.validAPITokenMultiSdkKey.Raw)
	req.Header.Add(OptlySDKHeader, "SDK_KEY_2")
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
		Clients:     make([]config.OAuthClientCredentials, 0),
		HMACSecrets: suite.signatures,
		TTL:         0,
		JwksURL:     "fake_url",
	}

	auth := NewAuth(authConfig)
	suite.Nil(auth)
}

func TestAuth(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}

func TestAuthValidCheckTokenFromJwksURLwithUpdating(t *testing.T) {

	const tk = `eyJhbGciOiJSUzI1NiIsImtpZCI6Il9ZdXhXVHgyZHAyRVNVb2s3MmUzcjNLb0R6OWZueFdJM29DQndOYnkyX0UiLCJ0eXAiOiJKV1QifQ.eyJhY2NvdW50X2lkIjo0Njg1MjgwNDQ0LCJhdWQiOiJTUlZDIiwiZXhwIjoxNTgyMzI1NjE5LCJpYXQiOjE1ODIzMjUwMTksImlzcyI6IlRPS0VOX1NFUlZJQ0UiLCJqdGkiOiI2OWVlN2M2NS1jNWU1LTQwNzYtYjI2Zi0yOGYzY2JlZjQwZjUiLCJuYmYiOjE1ODIzMjUwMTksInByb2plY3RfaWQiOjkyNjQzNjc2OTAsInNjb3BlcyI6ImF0dHJpYnV0ZXMubW9kaWZ5IGF0dHJpYnV0ZXMucmVhZCBhdWRpZW5jZXMucmVhZCBjaGFuZ2VfaGlzdG9yeS5yZWFkIGNvbGxhYm9yYXRvcnMubW9kaWZ5IGNvbGxhYm9yYXRvcnMucmVhZCBkY3AubW9kaWZ5IGRjcC5yZWFkIGV2ZW50cy5yZWFkIGV4cGVyaW1lbnRzLm1vZGlmeSBleHBlcmltZW50cy5yZWFkIGV4dGVuc2lvbnMubW9kaWZ5IGV4dGVuc2lvbnMucmVhZCBwYWdlcy5yZWFkIHByb2plY3RzLnJlYWQgcmVjb21tZW5kZXJzLm1vZGlmeSByZWNvbW1lbmRlcnMucmVhZCByZXN1bHRzLnJlYWQgc2FyLm1vZGlmeSBzYXIucmVhZCB1c2VyLnJlYWQiLCJzdWIiOiJ1cm46dXNlcjplMDk4ZjYwMGMwNzkxMWU1YjQ0NmU1MGRmMzdhOWEyMiIsInVzZXJfaWQiOiJlMDk4ZjYwMGMwNzkxMWU1YjQ0NmU1MGRmMzdhOWEyMiJ9.D9KVOiyvMP8ctJHIJAjt1ddj4Dol1c7vmPc0ZJg9A7t-yOv3WDjlxYMeTOPwPvN3iTHxIb-MFGIQyDpv63v13s00G0P4CFJHdXYBYTQETHCH1kFfjU5hK1lUAlqel3v25-uE-LgOnpnDsJK_LBmPwGJxh1_S5lyY6fBpQo9guMgmFIoN-GXGHzSWMD93oyD5CoiXWbxvLMIGMOrafl3YzqnEPK4WgmujnSR2vnj5lSLuJF_5-EICSXwuK2JVOq0xjGwa2trhw6xeVzN7JcKMb_baRq2tKxiiOjTnC-jPtkR22G8CWFcWUtOkkl-9XM9PXop2tHyLDWXxk73RChpAHg`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.String() == "/good" {
			fmt.Fprintln(w, `{"keys":[{"alg":"RS256","e":"AQAB","kid":"_YuxWTx2dp2ESUok72e3r3KoDz9fnxWI3oCBwNby2_E","kty":"RSA","n":"2ZNUw2VOO30mR15JcT5Lz85GznV2p3K0DtXRJiOhGOD0YnCkNZL3cPHR_r7_eVVJMokz4yGIW8hwSJN0GrzmihzULDTpFlAmSkSissSMIYANZOdHOPm5iCYCCeX_5ceCtDS85Z2gh0dN7vX7GkoYxJs-eLc0W8EVzA5V8S9c42ARGenH99nX8CiwiINEoZyLvv-Le2RX5zetVWqVD6EfP-mjzku-h5Nxx4PLk8tdiSpV-DllVGoYt5_P9_FgyTsZ1-62e2GJmNy0odZEUsTAxWnF_c1InEQZggI-vtCPNNVF1qgjArc86mGBc6z26EmRU91TavehP6n_oszhif83QQ","use":"sig"}]}`)
		}
		if r.URL.String() == "/bad" {
			fmt.Fprintln(w, `{"keys":[{"alg":"RS256","e":"AQAB","kid":"bad_id","kty":"RSA","n":"2ZNUw2VOO30mR15JcT5Lz85GznV2p3K0DtXRJiOhGOD0YnCkNZL3cPHR_r7_eVVJMokz4yGIW8hwSJN0GrzmihzULDTpFlAmSkSissSMIYANZOdHOPm5iCYCCeX_5ceCtDS85Z2gh0dN7vX7GkoYxJs-eLc0W8EVzA5V8S9c42ARGenH99nX8CiwiINEoZyLvv-Le2RX5zetVWqVD6EfP-mjzku-h5Nxx4PLk8tdiSpV-DllVGoYt5_P9_FgyTsZ1-62e2GJmNy0odZEUsTAxWnF_c1InEQZggI-vtCPNNVF1qgjArc86mGBc6z26EmRU91TavehP6n_oszhif83QQ","use":"sig"}]}`)
		}
	}))

	defer server.Close()

	validJwksURL := server.URL + "/good"
	invalidJwksURL := server.URL + "/bad"

	// constructing witih skipping claims validation
	verifier := &JWTVerifierURL{jwksURL: "fake_url", parser: &jwt.Parser{SkipClaimsValidation: true}}
	verifier.updateKeySet()
	go verifier.startTicker(time.Second)

	auth := NewAuth(&config.ServiceAuthConfig{})

	auth.Verifier = verifier

	token, err := auth.CheckToken(tk) // using fake_url
	assert.Nil(t, token)
	assert.Error(t, err)

	verifier.jwksLock.Lock()
	verifier.jwksURL = validJwksURL
	verifier.jwksLock.Unlock()

	token, err = auth.CheckToken(tk)
	assert.Nil(t, token) // still bad - polling every minute, still using fake_url
	assert.Error(t, err)

	<-time.After(1200 * time.Millisecond)

	token, err = auth.CheckToken(tk)
	assert.Equal(t, tk, token.Raw) // using /good URL
	assert.NoError(t, err)

	verifier.jwksLock.Lock()
	verifier.jwksURL = invalidJwksURL
	verifier.jwksLock.Unlock()
	<-time.After(1200 * time.Millisecond)

	token, err = auth.CheckToken(tk)
	assert.Nil(t, token) // changed to bad, after a minute using /bad URL
	assert.Error(t, err)
}
