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

// generate_secret is a tool for generating random secrets and their associated hashes.
// For more information, see docs/auth.md.

// package main
package main

import (
	"fmt"
	"github.com/optimizely/agent/pkg/jwtauth"
)

func main() {
	secretStr, hashStr, err := jwtauth.GenerateClientSecretAndHash()
	if err != nil {
		fmt.Printf("error: %v\n", err)
	} else {
		fmt.Printf("Client Secret: %v\n", secretStr)
		fmt.Printf("Client Secret's hash: %v\n", hashStr)
	}
}
