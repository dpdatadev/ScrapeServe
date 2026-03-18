package main

import (
	"errors"
	"log"
	"os"
)

// Return files for Logging or dumping
func GetFileWrite(fileName string) *os.File {
	if fileName == "" {
		log.Fatalf("errors.New(\"\"): %v\n", errors.New("WRITE FILE ERROR"))
		return nil
	}

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("errors.New(\"\"): %v\n", err)
		return nil
	}

	return file
}
