package logger

import (
	"bufio"
	"io"
	"log"
)

func Stream(reader io.ReadCloser, prefix string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		log.Printf("[%s]: %s", prefix, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading stream: %v", err)
	}
}
