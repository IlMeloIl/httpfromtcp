package headers

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	crlf = "\r\n"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))

	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		return 2, true, nil
	}

	line := string(data[:idx])
	parts := strings.SplitN(line, ":", 2)

	if len(parts) != 2 {
		return 0, false, fmt.Errorf("malformed header: %s", line)
	}

	rawKey := parts[0]
	if strings.TrimRight(rawKey, " \t") != rawKey {
		return 0, false, fmt.Errorf("spaces between header name and colon not allowed")
	}

	key := strings.ToLower(strings.TrimSpace(parts[0]))
	value := strings.TrimSpace(parts[1])

	if key == "" {
		return 0, false, fmt.Errorf("empty header key not allowed")
	}

	if !validateHeaderKey(key) {
		return 0, false, fmt.Errorf("invalid characters in header name: %s", key)
	}

	if existingValue, ok := h[key]; !ok {
		h[key] = value
	} else {
		h[key] = fmt.Sprintf("%s, %s", existingValue, value)
	}

	return idx + len(crlf), false, nil
}

func validateHeaderKey(key string) bool {
	for _, char := range key {
		isAlpha := (char >= 'a' && char <= 'z')
		isDigit := (char >= '0' && char <= '9')
		isSpecial := strings.ContainsRune("!#$%&'*+-.^_`|~", char)

		if !isAlpha && !isDigit && !isSpecial {
			return false
		}
	}
	return true
}

func (h Headers) Get(key string) string {
	value, ok := h[key]
	if !ok {
		return ""
	}

	return value
}

func (h Headers) Set(key, value string) {
	h[strings.ToLower(key)] = value
}
