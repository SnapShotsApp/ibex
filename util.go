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
