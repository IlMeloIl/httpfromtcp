package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)

	go func() {
		defer close(lines)
		defer f.Close()

		var currentLine string

		for {
			b := make([]byte, 8)
			_, err := f.Read(b)
			if err != nil {
				if errors.Is(err, io.EOF) {
					if currentLine != "" {
						lines <- currentLine
					}
					break
				}
				log.Printf("error: %s\n", err.Error())
				return
			}

			parts := strings.Split(string(b), "\n")

			for i := 0; i < len(parts)-1; i++ {
				currentLine += parts[i]
				lines <- currentLine
				currentLine = ""
			}
			currentLine += parts[len(parts)-1]
		}
	}()
	return lines
}

func main() {
	f, err := os.Open("messages.txt")
	if err != nil {
		log.Fatal(err)
	}

	lines := getLinesChannel(f)

	for line := range lines {
		fmt.Printf("read: %s\n", line)
	}
}
