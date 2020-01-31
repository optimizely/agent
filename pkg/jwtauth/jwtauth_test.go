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

// Package jwtauth contains JWT-related helpers
package jwtauth

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type JWTAuthTestSuite struct{
	suite.Suite
}

func (s *JWTAuthTestSuite) TestBuildAPIAccessTokenSuccess() {
	tokenTtl := 10 * time.Minute
	secretKey := []byte("seekrit")
	tokenString, expiresIn, err := BuildAPIAccessToken("123", tokenTtl, secretKey)
	s.NoError(err)
	expectedExpiresIn := time.Now().Add(tokenTtl).Unix()
	s.GreaterOrEqual(expectedExpiresIn, expiresIn)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (i interface{}, err error) {
		return secretKey, nil
	})
	s.NoError(err)
	s.True(token.Valid)
	claims, ok := token.Claims.(jwt.MapClaims)
	s.True(ok)
	s.Equal("123", claims["sdk_key"])
	claimsExpFloat, ok := claims["exp"].(float64)
	s.True(ok)
	s.Equal(expiresIn, int64(claimsExpFloat))
}

func (s *JWTAuthTestSuite) TestBuildAdminAccessTokenSuccess() {
	tokenTtl := 10 * time.Minute
	secretKey := []byte("seekrit")
	tokenString, expiresIn, err := BuildAdminAccessToken(tokenTtl, secretKey)
	s.NoError(err)
	expectedExpiresIn := time.Now().Add(tokenTtl).Unix()
	s.GreaterOrEqual(expectedExpiresIn, expiresIn)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (i interface{}, err error) {
		return secretKey, nil
	})
	s.NoError(err)
	s.True(token.Valid)
	claims, ok := token.Claims.(jwt.MapClaims)
	s.True(ok)
	s.Equal(true, claims["admin"])
	claimsExpFloat, ok := claims["exp"].(float64)
	s.True(ok)
	s.Equal(expiresIn, int64(claimsExpFloat))
}

func TestJWTAuthTestSuite(t *testing.T) {
	suite.Run(t, new(JWTAuthTestSuite))
}
