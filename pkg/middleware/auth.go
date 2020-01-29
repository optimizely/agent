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

	"github.com/dgrijalva/jwt-go"
	"github.com/rs/zerolog/log"
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

	checkClaims map[string]struct{}
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
		log.Print("rejected, token", token, err)
		return nil, err
	}

	if !tk.Valid {
		return nil, errors.New("invalid token")
	}

	return tk, nil
}

// Verify gets string token from the requst and validates it
func (a Auth) Verify(r *http.Request) (*jwt.Token, error) {

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

// Authorize is middleware for auth
func (a Auth) Authorize(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		tk, err := a.Verify(r)

		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "unauthorized, "reason": "%v"}`, err), http.StatusUnauthorized)
			return
		}

		if a.enabled() {
			claims := tk.Claims.(jwt.MapClaims)

			for key := range a.checkClaims {

				value, ok := claims[key]
				if !ok {
					http.Error(w, fmt.Sprintf(`{"error": "unauthorized, "reason": "%s key not in claims"}`, key), http.StatusUnauthorized)
					return
				}

				switch key {
				case "exp":
					if expired := (getNumberFromJSON(value) - time.Now().Unix()) <= 0; expired {
						http.Error(w, `{"error": "unauthorized", "reason": "token expired"}`, http.StatusUnauthorized)
						return
					}

				case "sdk_key":
					sdkKeyFromHeader := r.Header.Get(OptlySDKHeader)
					if sdkKey, ok := value.(string); !ok || sdkKey != sdkKeyFromHeader {
						http.Error(w, `{"error": "unauthorized", "reason": "SDK keys not equal"}`, http.StatusUnauthorized)
						return
					}
				case "admin":
					if adminFlag, ok := value.(bool); !ok || !adminFlag {
						http.Error(w, `{"error": "unauthorized", "reason": "admin flag not set"}`, http.StatusUnauthorized)
						return
					}
				}
			}
		}

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// NewAuth makes Auth middleware
func NewAuth(v Verifier, checkClaims map[string]struct{}) Auth {
	return Auth{Verifier: v, checkClaims: checkClaims}
}
