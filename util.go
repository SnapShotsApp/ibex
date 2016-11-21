/* Copyright 2016 Snapshots LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"io"
	"os"
)

func closeQuietly(handle io.Closer) {
	err := handle.Close()
	if err != nil {
		Warn("Error on close: %s", err)
	}
}

func handleErr(err error) {
	if err != nil {
		Fatal("Uncaught error: %s", err)
	}

}

func mustFetchEnv(v string) string {
	envVar := os.Getenv(v)

	if len(envVar) <= 0 {
		Fatal("Required environment variable %s not found", v)
	}

	return envVar
}
