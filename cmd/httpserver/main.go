package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	response "httpfromtcp/internal/respose"
	server "httpfromtcp/internal/server"
)

const port = (42069)

func proxyHandler(w *response.Writer, targetPath string) error {
	targetUrl := "https://httpbin.org" + targetPath

	resp, err := http.Get(targetUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	w.WriteStatusLine(response.StatusCode(resp.StatusCode))

	h := headers.NewHeaders()

	for key, values := range resp.Header {
		if strings.ToLower(key) != "content-length" {
			h.Set(key, strings.Join(values, ","))
		}
	}

	h.Set("transfer-encoding", "chunked")

	h.Set("trailer", "X-Content-SHA256, X-Content-Length")

	err = w.WriteHeaders(h)
	if err != nil {
		return err
	}

	var fullResponseBody bytes.Buffer

	buffer := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {

			fullResponseBody.Write(buffer[:n])

			_, err := w.WriteChunkedBody(buffer[:n])
			if err != nil {
				return err
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		return err
	}

	rawBody := fullResponseBody.Bytes()
	bodyHash := sha256.Sum256(rawBody)
	hashHex := hex.EncodeToString(bodyHash[:])

	trailers := headers.NewHeaders()
	trailers.Set("X-Content-SHA256", hashHex)
	trailers.Set("X-Content-Length", fmt.Sprintf("%d", fullResponseBody.Len()))

	return w.WriteTrailers(trailers)
}

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
		if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
			targetPath := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
			if targetPath == "" {
				targetPath = "/"
			}

			err := proxyHandler(w, targetPath)
			if err != nil {
				log.Printf("Error proxying request: %v", err)

				if w.IsInitialized() {
					w.WriteStatusLine(response.StatusInternalServerError)
					h := headers.NewHeaders()
					h.Set("content-type", "text/html")
					h.Set("content-length", fmt.Sprintf("%d", len(internalErrorHTML)))
					w.WriteHeaders(h)
					w.WriteBody([]byte(internalErrorHTML))

				}
			}
			return
		}

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

		case "/video":
			videoData, err := os.ReadFile("../../assets/vim.mp4")
			if err != nil {
				w.WriteStatusLine(response.StatusInternalServerError)
				h.Set("content-type", "text/html")
				h.Set("content-length", fmt.Sprintf("%d", len(internalErrorHTML)))
				w.WriteHeaders(h)
				w.WriteBody([]byte(internalErrorHTML))
				return
			}

			w.WriteStatusLine(response.StatusOK)
			h.Set("content-type", "video/mp4")
			h.Set("content-length", fmt.Sprintf("%d", len(videoData)))
			w.WriteHeaders(h)
			w.WriteBody(videoData)

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
