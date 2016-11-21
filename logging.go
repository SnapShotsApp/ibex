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
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
)

var (
	debugLogger *log.Logger
	infoLogger  *log.Logger
	errLogger   *log.Logger
)

func init() {
	flags := log.LstdFlags | log.LUTC
	debugLogger = log.New(os.Stdout, "", flags)
	infoLogger = log.New(os.Stdout, "", flags)
	errLogger = log.New(os.Stderr, "", flags)
}

// Info logs a statement to the log
func Info(format string, vars ...interface{}) {
	infoLogger.SetPrefix(getPrefixStr("[INFO]"))
	infoLogger.Printf(format, vars...)
}

// Warn logs a warning statement to stderr
func Warn(format string, vars ...interface{}) {
	errLogger.SetPrefix(getPrefixStr("[WARN]"))
	errLogger.Printf(format, vars...)

}

// Debug logs a statement to the log if verbosity is turned on
func Debug(format string, vars ...interface{}) {
	if debugLogEnabled {
		debugLogger.SetPrefix(getPrefixStr("[DEBUG]"))
		debugLogger.Printf(format, vars...)
	}
}

// Fatal logs a statement to the log and then panics
func Fatal(format string, vars ...interface{}) {
	errLogger.SetPrefix(getPrefixStr("[FATAL]"))
	errLogger.Fatalf(format, vars...)
}

func getPrefixStr(prefix string) string {
	_, file, line, ok := runtime.Caller(2)
	if ok {
		basePath := path.Clean(path.Base(file))
		return fmt.Sprintf("%s\t%s:%d\t", prefix, basePath, line)
	}

	return fmt.Sprintf("%s\t\t", prefix)
}
