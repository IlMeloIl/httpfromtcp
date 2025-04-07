package server

import (
	"fmt"
	"httpfromtcp/internal/request"
	response "httpfromtcp/internal/respose"
	"log"
	"net"
	"sync/atomic"
)

type Server struct {
	listener net.Listener
	isClosed atomic.Bool
	port     int
	handler  Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}
	server := &Server{
		listener: ln,
		port:     port,
		handler:  handler,
	}

	go server.listen()

	return server, nil
}

func (s *Server) Close() error {
	s.isClosed.Store(true)

	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()

		if s.isClosed.Load() {
			return
		}

		if err != nil {
			log.Printf("error accepting connection: %v", err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Printf("Error parsing request: %v", err)

		w := response.NewWriter(conn)

		w.WriteStatusLine(response.StatusBadRequest)

		errMsg := "Bad Request\n"
		headers := response.GetDefaultHeaders(len(errMsg))
		w.WriteHeaders(headers)

		w.WriteBody([]byte(errMsg))
		return
	}

	w := response.NewWriter(conn)

	s.handler(w, req)
}
