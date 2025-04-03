package request

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	h "httpfromtcp/internal/headers"
)

const (
	StateInitialized = iota
	StateParsingHeaders
	StateParsingBody
	StateDone
)

type Request struct {
	RequestLine RequestLine
	Headers     h.Headers
	Body        []byte
	state       int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const (
	crlf       = "\r\n"
	bufferSize = 8
)

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(crlf))

	if idx == -1 {
		return nil, 0, nil
	}

	requestLineText := string(data[:idx])

	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, err
	}

	return requestLine, idx + len(crlf), nil
}

func requestLineFromString(str string) (*RequestLine, error) {
	parts := strings.Split(str, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("poorly formatted request-line: %s", str)
	}

	method := parts[0]
	for _, c := range method {
		if c < 'A' || c > 'Z' {
			return nil, fmt.Errorf("invalid method: %s", method)
		}
	}

	requestTarget := parts[1]

	versionParts := strings.Split(parts[2], "/")
	if len(versionParts) != 2 {
		return nil, fmt.Errorf("malformed start-line: %s", str)
	}

	httpPart := versionParts[0]
	if httpPart != "HTTP" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", httpPart)
	}

	version := versionParts[1]
	if version != "1.1" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", version)
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   versionParts[1],
	}, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0

	for r.state != StateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return totalBytesParsed, err
		}

		if n == 0 {
			break
		}

		totalBytesParsed += n
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case StateInitialized:
		requestLine, bytesConsumed, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}

		if bytesConsumed == 0 {
			return 0, nil
		}

		r.RequestLine = *requestLine
		r.state = StateParsingHeaders

		if r.Headers == nil {
			r.Headers = h.NewHeaders()
		}

		return bytesConsumed, nil

	case StateParsingHeaders:
		if r.Headers == nil {
			r.Headers = h.NewHeaders()
		}

		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}

		if n == 0 {
			return 0, nil
		}

		if done {
			r.state = StateParsingBody
		}

		return n, nil

	case StateParsingBody:
		if r.Headers == nil {
			return 0, nil
		}

		contentLengthStr := r.Headers.Get("content-length")
		if contentLengthStr == "" {
			r.state = StateDone
			return 0, nil
		}

		contentLength, err := strconv.Atoi(contentLengthStr)
		if err != nil {
			return 0, fmt.Errorf("error converting contentLengthStr to integer: %w", err)
		}

		if contentLength == 0 {
			r.state = StateDone
			return 0, nil
		}

		if r.Body == nil {
			r.Body = make([]byte, 0, contentLength)
		}

		bytesToConsume := len(data)
		if len(r.Body)+bytesToConsume > contentLength {
			bytesToConsume = contentLength - len(r.Body)
		}

		r.Body = append(r.Body, data[:bytesToConsume]...)

		if len(r.Body) == contentLength {
			r.state = StateDone
		}
		return bytesToConsume, nil

	case StateDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")

	default:
		return 0, fmt.Errorf("error: unknown state")
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0

	request := &Request{
		state: StateInitialized,
	}

	for request.state != StateDone {

		if readToIndex == len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIndex:])

		readToIndex += n

		if err == io.EOF {

			if readToIndex > 0 {
				bytesConsumed, parseErr := request.parse(buf[:readToIndex])
				if parseErr != nil {
					return nil, parseErr
				}

				if bytesConsumed == 0 {
					if request.state == StateParsingBody {
						contentLengthStr := request.Headers.Get("content-length")
						if contentLengthStr != "" {
							contentLength, _ := strconv.Atoi(contentLengthStr)
							if contentLength > 0 && len(request.Body) < contentLength {
								return nil, fmt.Errorf("unexpected EOF: incomplete request")
							}
						}
					} else {
						return nil, fmt.Errorf("unexpected EOF: incomplete request")
					}
				}
			}
			break
		} else if err != nil {
			return nil, err
		}

		bytesConsumed, parseErr := request.parse(buf[:readToIndex])
		if parseErr != nil {
			return nil, parseErr
		}

		if bytesConsumed > 0 {
			remaining := readToIndex - bytesConsumed
			copy(buf, buf[bytesConsumed:readToIndex])
			readToIndex = remaining
		}
	}

	if request.state != StateDone {
		return nil, fmt.Errorf("unexpected EOF: incomplete request")
	}

	return request, nil
}
