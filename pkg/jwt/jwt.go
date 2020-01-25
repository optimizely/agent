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

// package jwt contains JWT-related helpers
package jwt

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
)

// BuildAPIAccessToken returns a token for accessing the API service using the argument SDK key and TTL. It also returns the expiration timestamp.
func BuildAPIAccessToken(sdkKey string, ttl time.Duration, key []byte) (string, int64, error) {
	expires := time.Now().Add(ttl).Unix()
	// TODO: should use any of these standards claims? https://tools.ietf.org/html/rfc7519#section-4.1
	// Should change "expires" to "exp" to match the standard one?
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sdk_key": sdkKey,
		"expires": expires,
	})
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", 0, fmt.Errorf("error building API access token: %w", err)
	}
	return tokenString, expires, nil
}
