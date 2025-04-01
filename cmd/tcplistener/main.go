package main

import (
	"fmt"
	"log"
	"net"

	r "httpfromtcp/internal/request"
)

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

		req, err := r.RequestFromReader(conn)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Request line:\n- Method: %v\n- Target: %v\n- Version: %v\n", req.RequestLine.Method, req.RequestLine.RequestTarget, req.RequestLine.HttpVersion)
		fmt.Printf("Headers:\n")
		for key, val := range req.Headers {
			fmt.Printf("- %s: %s\n", key, val)
		}

		fmt.Println("Connection closed")
	}

}
