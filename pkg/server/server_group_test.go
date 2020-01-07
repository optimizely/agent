/****************************************************************************
 * Copyright 2019, Optimizely, Inc. and contributors                        *
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

package server

import (
	"context"
	"testing"
)

func TestServeAndShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	sg := NewGroup(ctx, conf)

	sg.GoListenAndServe("valid1", "1000", handler)
	sg.GoListenAndServe("valid2", "1001", handler)

	cancel()
	sg.Wait()
}

func TestInvalidServer(t *testing.T) {
	sg := NewGroup(context.Background(), conf)
	sg.GoListenAndServe("invalid", "-1", handler)
	sg.Wait() // Don't need to shutdown since server never started
}
