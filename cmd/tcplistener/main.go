package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func getLinesChannel(conn net.Conn) <-chan string {
	lines := make(chan string)

	go func() {
		defer close(lines)
		defer conn.Close()

		var currentLine string

		for {
			b := make([]byte, 8)
			n, err := conn.Read(b)
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

			if n == 0 {
				continue
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

const Port = "42069"

func main() {
	ln, err := net.Listen("tcp", ":"+Port)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Listening to TCP connections on port %s ...\n", Port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Could not accept conn: %s\n", err)
			continue
		}

		fmt.Println("Connection accepted!")

		lines := getLinesChannel(conn)

		for line := range lines {
			fmt.Println(line)
		}
		fmt.Println("Connection closed")
	}

}
