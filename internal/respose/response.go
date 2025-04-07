package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
)

type WriterState int

const (
	StateInitialized WriterState = iota
	StateStatusWritten
	StateHeadersWritten
	StateBodyWriten
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

type Writer struct {
	w     io.Writer
	state WriterState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w:     w,
		state: StateInitialized,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != StateInitialized {
		return fmt.Errorf("status line already written or called out of order")
	}
	var reasonPhrase string

	switch statusCode {
	case StatusOK:
		reasonPhrase = "OK"
	case StatusBadRequest:
		reasonPhrase = "Bad Request"
	case StatusInternalServerError:
		reasonPhrase = "Internal Server Error"
	}

	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase)
	_, err := w.w.Write([]byte(statusLine))
	if err != nil {
		return err
	}

	w.state = StateStatusWritten
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h["content-length"] = fmt.Sprintf("%d", contentLen)
	h["connection"] = "close"
	h["content-type"] = "text/plain"

	return h
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != StateStatusWritten {
		return fmt.Errorf("status line not written or headers already written")
	}
	for k, v := range headers {
		headerLine := fmt.Sprintf("%s: %s\r\n", k, v)
		_, err := w.w.Write([]byte(headerLine))
		if err != nil {
			return err
		}
	}
	_, err := w.w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	w.state = StateHeadersWritten
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != StateHeadersWritten {
		return 0, fmt.Errorf("headers not written or body already written")
	}

	n, err := w.w.Write(p)
	if err != nil {
		return 0, err
	}

	w.state = StateBodyWriten
	return n, nil
}
