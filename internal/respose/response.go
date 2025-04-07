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
	StateBodyWritten
	StateChunkedBodyWriting
	StateTrailersWritten
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

	w.state = StateBodyWritten
	return n, nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.state != StateHeadersWritten && w.state != StateChunkedBodyWriting {
		return 0, fmt.Errorf("headers not written or non-chunked body already written")
	}

	w.state = StateChunkedBodyWriting

	chunkSizeHex := fmt.Sprintf("%x\r\n", len(p))
	_, err := w.w.Write([]byte(chunkSizeHex))
	if err != nil {
		return 0, err
	}

	n, err := w.w.Write(p)
	if err != nil {
		return 0, err
	}

	_, err = w.w.Write([]byte("\r\n"))
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.state != StateChunkedBodyWriting {
		return 0, fmt.Errorf("not in chunked body writing state")
	}

	_, err := w.w.Write([]byte("0\r\n"))
	if err != nil {
		return 0, err
	}

	w.state = StateBodyWritten
	return 0, nil
}

func (w *Writer) IsInitialized() bool {
	return w.state == StateInitialized
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.state != StateBodyWritten {
		return fmt.Errorf("trailers can only be written after a chunked body is complete")
	}

	for k, v := range h {
		trailerLine := fmt.Sprintf("%s: %s\r\n", k, v)
		_, err := w.w.Write([]byte(trailerLine))
		if err != nil {
			return err
		}
	}

	_, err := w.w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	return nil
}
