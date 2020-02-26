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

package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

var secretVersion int = 1
var bcryptWorkFactor int = 12

func main() {
	randBytes := make([]byte, 32)
	_, err := rand.Read(randBytes)
	if err != nil {
		fmt.Println("error returned from rand.Read:", err)
		return
	}

	encoded := base64.StdEncoding.EncodeToString(randBytes)
	stripped := strings.TrimSuffix(encoded, "=")
	secretStr := fmt.Sprintf("%v:%v", secretVersion, stripped)

	hashBytes, err := bcrypt.GenerateFromPassword(randBytes, 12)
	if err != nil {
		fmt.Println("error returned from bcrypt.GenerateFromPassword:", err)
		return
	}
	hashStr := base64.StdEncoding.EncodeToString(hashBytes)

	fmt.Printf("Client Secret: %v\n", secretStr)
	fmt.Printf("Client Secret's hash: %v\n", hashStr)
}
