package server

import (
	"httpfromtcp/internal/request"
	response "httpfromtcp/internal/respose"
)

type Handler func(w *response.Writer, req *request.Request)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}
