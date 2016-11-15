package main

import (
	"io"
	"log"
	"os"
)

func closeQuietly(handle io.Closer) {
	err := handle.Close()
	if err != nil {
		log.Printf("Warning: %s", err)
	}
}

func handleErr(err error) {
	if err != nil {
		log.Fatal(err)
	}

}

func mustFetchEnv(v string) string {
	envVar := os.Getenv(v)

	if len(envVar) <= 0 {
		log.Fatalf("Required environment variable %s not found", v)
	}

	return envVar
}
