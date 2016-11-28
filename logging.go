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
	"io"
	"log"
	"os"
	"path"
	"runtime"
)

const (
	logFlags = log.LstdFlags | log.LUTC
)

// Infoer defines an interface for logging Info-level messages
type Infoer interface {
	Info(string, ...interface{})
}

// Warner defines an interface for logging Warn-level messages
type Warner interface {
	Warn(string, ...interface{})
}

// Debugger defines an interface for logging Debug-level messages
type Debugger interface {
	Debug(string, ...interface{})
}

// Fataler defines an interface for logging Fatal-level messages
type Fataler interface {
	Fatal(string, ...interface{})
}

// ErrorHandler defines an interface for processing errors through
// the logging system
type ErrorHandler interface {
	HandleErr(interface{})
	CloseQuietly(io.Closer)
}

// ILogger combines the other logging interfaces
type ILogger interface {
	Infoer
	Warner
	Debugger
	Fataler
	ErrorHandler
}

type ibexLogger struct {
	debug       bool
	debugLogger *log.Logger
	infoLogger  *log.Logger
	errLogger   *log.Logger
}

func newLogger(debug bool) ibexLogger {
	logger := ibexLogger{
		debug,
		log.New(os.Stdout, "", logFlags),
		log.New(os.Stdout, "", logFlags),
		log.New(os.Stderr, "", logFlags),
	}

	return logger
}

// Info logs a statement to the log
func (l ibexLogger) Info(format string, vars ...interface{}) {
	l.infoLogger.SetPrefix(getPrefixStr("[INFO]"))
	l.infoLogger.Printf(format, vars...)
}

// Warn logs a warning statement to stderr
func (l ibexLogger) Warn(format string, vars ...interface{}) {
	l.errLogger.SetPrefix(getPrefixStr("[WARN]"))
	l.errLogger.Printf(format, vars...)

}

// Debug logs a statement to the log if verbosity is turned on
func (l ibexLogger) Debug(format string, vars ...interface{}) {
	if l.debug {
		l.debugLogger.SetPrefix(getPrefixStr("[DEBUG]"))
		l.debugLogger.Printf(format, vars...)
	}
}

// Fatal logs a statement to the log and then panics
func (l ibexLogger) Fatal(format string, vars ...interface{}) {
	l.errLogger.SetPrefix(getPrefixStr("[FATAL]"))
	l.errLogger.Fatalf(format, vars...)
}

func (l ibexLogger) HandleErr(err interface{}) {
	if err == nil {
		return
	}

	switch err := err.(type) {
	case string:
		l.Fatal(err)
	case error:
		l.Fatal(err.Error())
	default:
		l.Fatal("Received fatal error of unknown type: %v", err)
	}
}

func (l ibexLogger) CloseQuietly(handle io.Closer) {
	err := handle.Close()
	l.HandleErr(err)
}

func getPrefixStr(prefix string) string {
	_, file, line, ok := runtime.Caller(2)
	if ok {
		basePath := path.Clean(path.Base(file))
		return fmt.Sprintf("%s\t%s:%d\t", prefix, basePath, line)
	}

	return fmt.Sprintf("%s\t\t", prefix)
}
