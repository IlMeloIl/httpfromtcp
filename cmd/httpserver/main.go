package main

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"log"
	"os"
	"os/signal"
	"syscall"

	response "httpfromtcp/internal/respose"
	server "httpfromtcp/internal/server"
)

const port = (42069)

func main() {
	badRequestHTML := `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`

	internalErrorHTML := `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`

	successHTML := `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`

	handler := func(w *response.Writer, req *request.Request) {
		h := headers.NewHeaders()
		h.Set("content-type", "text/html")
		h.Set("connection", "close")

		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			w.WriteStatusLine(response.StatusBadRequest)

			h.Set("content-length", fmt.Sprintf("%d", len(badRequestHTML)))
			w.WriteHeaders(h)
			w.WriteBody([]byte(badRequestHTML))

		case "/myproblem":
			w.WriteStatusLine(response.StatusInternalServerError)

			h.Set("content-length", fmt.Sprintf("%d", len(internalErrorHTML)))
			w.WriteHeaders(h)
			w.WriteBody([]byte(internalErrorHTML))
		default:
			w.WriteStatusLine(response.StatusOK)

			h.Set("content-length", fmt.Sprintf("%d", len(successHTML)))
			w.WriteHeaders(h)
			w.WriteBody([]byte(successHTML))
		}
	}

	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting the server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port:", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
