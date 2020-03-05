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

var secretBytesLen int = 32

var bcryptWorkFactor int = 12

// ValidateClientSecret returns true if the hash of the secret provided in config matches
// the secret provided in the request. Returns an error if the req secret fails base64
// decoding.
func ValidateClientSecret(reqSecret string, configSecretHash []byte) (bool, error) {
	secretBytes, err := base64.StdEncoding.DecodeString(reqSecret)
	if err != nil {
		return false, fmt.Errorf("error decoding string: %v", err)
	}
	return bcrypt.CompareHashAndPassword(configSecretHash, secretBytes) == nil, nil
}

// GenerateClientSecretAndHash returns a random secret and its hash, for use with
// Agent's authN/authZ workflow.
// - The first return value is the secret - 32 random bytes, base64-encoded.
// - The second return value is the bcrypt hash of the secret.
// - The hash should be included in Agent's auth configuration as the client_secret value.
// - The secret should be sent in the request to the token issuer endpoint.
func GenerateClientSecretAndHash() (string, string, error) {
	secretBytes := make([]byte, secretBytesLen)
	_, err := rand.Read(secretBytes)
	if err != nil {
		return "", "", fmt.Errorf("error returned from rand.Read: %v", err)
	}
	secretStr := base64.StdEncoding.EncodeToString(secretBytes)

	hashBytes, err := bcrypt.GenerateFromPassword(secretBytes, bcryptWorkFactor)
	if err != nil {
		return "", "", fmt.Errorf("error returned from bcrypt.GenerateFromPassword: %v", err)
	}
	hashStr := base64.StdEncoding.EncodeToString(hashBytes)

	return secretStr, hashStr, nil
}

// DecodeSecretHashFromConfig returns the decoded secret hash as a byte slice, or an error if decoding failed
func DecodeSecretHashFromConfig(configSecretHash string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(configSecretHash)
}
