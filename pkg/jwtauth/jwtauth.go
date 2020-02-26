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

// Package jwtauth contains JWT and authentication-related helpers
package jwtauth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// BuildAPIAccessToken returns a token for accessing the API service using the argument SDK key and TTL. It also returns the expiration timestamp.
func BuildAPIAccessToken(sdkKey string, ttl time.Duration, key []byte) (tokenString string, err error) {
	expires := time.Now().Add(ttl).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":     "Optimizely",
		"sdk_key": sdkKey,
		"exp":     expires,
	})
	tokenString, err = token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("error building API access token: %w", err)
	}
	return tokenString, nil
}

// BuildAdminAccessToken returns a token for accessing the Admin service using the argument TTL. It also returns the expiration timestamp.
func BuildAdminAccessToken(ttl time.Duration, key []byte) (tokenString string, err error) {
	expires := time.Now().Add(ttl).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":   "Optimizely",
		"exp":   expires,
		"admin": true,
	})
	tokenString, err = token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("error building Admin access token: %w", err)
	}
	return tokenString, nil
}

// ValidateClientSecret compares secret keys
func ValidateClientSecret(reqSecretStr string, configSecret []byte) bool {
	reqSecret := []byte(reqSecretStr)
	if len(configSecret) != len(reqSecret) {
		return false
	}
	return subtle.ConstantTimeCompare(reqSecret, configSecret) == 1
}

var secretBytesLen int = 32

var secretVersion int = 1

var bcryptWorkFactor int = 12

// GenerateClientSecretAndHash returns a random secret and its hash, for use with
// Agent's authN/authZ workflow.
// - The first return value is the secret, composed of a version number, separator,
//   and 32 random bytes, base64-encoded.
// - The second return value is the bcrypt hash of the random bytes from the secret.
// - The hash should be included in Agent's auth configuration as the client_secret value.
// - The secret should be sent in the request to the token issuer endpoint.
func GenerateClientSecretAndHash() (string, string, error) {
	secretBytes := make([]byte, secretBytesLen)
	_, err := rand.Read(secretBytes)
	if err != nil {
		return "", "", fmt.Errorf("error returned from rand.Read: %v", err)
	}

	encoded := base64.StdEncoding.EncodeToString(secretBytes)
	secretStr := fmt.Sprintf("%v:%v", secretVersion, encoded)

	hashBytes, err := bcrypt.GenerateFromPassword(secretBytes, 12)
	if err != nil {
		return "", "", fmt.Errorf("error returned from bcrypt.GenerateFromPassword: %v", err)
	}
	hashStr := base64.StdEncoding.EncodeToString(hashBytes)

	return secretStr, hashStr, nil
}
