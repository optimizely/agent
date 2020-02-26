/****************************************************************************
 * Copyright 2020, Optimizely, Inc. and contributors                   *
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
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/optimizely/agent/config"

	"github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"
)

func getNumberFromJSON(val interface{}) int64 {
	switch v := val.(type) {
	case int64:
		return v
	case float64:
		return int64(v)
	}
	return 0
}

// NoAuth is NoOp for auth
type NoAuth struct{}

// CheckToken returns no token and no error
func (NoAuth) CheckToken(string) (*jwt.Token, error) {
	return nil, nil
}

// Auth is the middleware for all REST API's
type Auth struct {
	Verifier
}

// Verifier checks token
type Verifier interface {
	CheckToken(string) (*jwt.Token, error)
}

// JWTVerifier checks token with JWT, implements Verifier
type JWTVerifier struct {
	secretKey string
}

// NewJWTVerifier creates JWTVerifier with secret key
func NewJWTVerifier(secretKey string) JWTVerifier {
	return JWTVerifier{secretKey: secretKey}
}

// CheckToken checks the token and returns it if it's valid
func (c JWTVerifier) CheckToken(token string) (*jwt.Token, error) {
	if token == "" {
		return nil, errors.New("invalid token")
	}

	tk, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(c.secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	if !tk.Valid {
		return nil, errors.New("invalid token")
	}

	return tk, nil
}

// JWTVerifierURL checks token with JWT against JWKS, implements Verifier
type JWTVerifierURL struct {
	jwksURL string

	parser *jwt.Parser
}

// NewJWTVerifierURL creates JWTVerifierURL with JWKS URL
func NewJWTVerifierURL(jwksURL string) JWTVerifierURL {
	return JWTVerifierURL{jwksURL: jwksURL, parser: new(jwt.Parser)}
}

// CheckToken checks the token, validates against JWKS and returns it if it's valid
func (c JWTVerifierURL) CheckToken(token string) (*jwt.Token, error) {
	if token == "" {
		return nil, errors.New("invalid token")
	}

	tk, err := c.parser.Parse(token, func(token *jwt.Token) (interface{}, error) {
		set, err := jwk.Fetch(c.jwksURL)
		if err != nil {
			return nil, err
		}

		keyID, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("expecting JWT header to have string kid")
		}

		if key := set.LookupKeyID(keyID); len(key) == 1 {
			return key[0].Materialize()
		}

		return nil, fmt.Errorf("unable to find key %q", keyID)
	})
	if err != nil {
		return nil, err
	}

	if !tk.Valid {
		return nil, errors.New("invalid token")
	}

	return tk, nil
}

func (a Auth) verify(r *http.Request) (*jwt.Token, error) {

	var token string

	if values, ok := r.Header["Auth"]; ok && len(values) > 0 {
		token = values[0]
	}

	if values, ok := r.Header["Jwt"]; ok && len(values) > 0 {
		token = values[0]
	}

	if values, ok := r.Header["Authorization"]; ok && len(values) > 0 {
		value := values[0]
		for _, key := range []string{"JWT", "Bearer"} {
			token = strings.TrimSpace(strings.TrimLeft(value, key))
		}
	}

	return a.CheckToken(token)

}

func (a Auth) enabled() bool {
	if _, ok := a.Verifier.(NoAuth); ok {
		return false
	}
	return true
}

// AuthorizeAdmin is middleware for admin auth
func (a Auth) AuthorizeAdmin(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		tk, err := a.verify(r)

		if err != nil {
			RenderError(err, http.StatusUnauthorized, w, r)
			return
		}

		if a.enabled() {
			claims := tk.Claims.(jwt.MapClaims)

			if expired := (getNumberFromJSON(claims["exp"]) - time.Now().Unix()) <= 0; expired {
				RenderError(errors.New("token expired"), http.StatusUnauthorized, w, r)
				return
			}
			if adminFlag, ok := claims["admin"].(bool); !ok || !adminFlag {
				RenderError(errors.New("admin flag not set"), http.StatusUnauthorized, w, r)
				return
			}
		}

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// AuthorizeAPI is middleware for auth api
func (a Auth) AuthorizeAPI(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		tk, err := a.verify(r)

		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "unauthorized, "reason": "%v"}`, err), http.StatusUnauthorized)
			return
		}

		if a.enabled() {
			claims := tk.Claims.(jwt.MapClaims)
			if expired := (getNumberFromJSON(claims["exp"]) - time.Now().Unix()) <= 0; expired {
				RenderError(errors.New("token expired"), http.StatusUnauthorized, w, r)
				return
			}
			sdkKeyFromHeader := r.Header.Get(OptlySDKHeader)
			if sdkKey, ok := claims["sdk_key"].(string); !ok || sdkKey != sdkKeyFromHeader {
				RenderError(errors.New("SDK keys not equal"), http.StatusUnauthorized, w, r)
				return
			}
		}

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// NewAuth makes Auth middleware
func NewAuth(authConfig *config.ServiceAuthConfig) Auth {

	if authConfig.JwksURL != "" {
		return Auth{Verifier: NewJWTVerifierURL(authConfig.JwksURL)}
	}

	if authConfig.HMACSecret == "" {
		return Auth{Verifier: NoAuth{}}
	}

	return Auth{Verifier: NewJWTVerifier(authConfig.HMACSecret)}

}
