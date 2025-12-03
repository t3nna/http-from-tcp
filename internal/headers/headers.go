package headers

import (
	"bytes"
	"fmt"
	"strings"
)

func isToken(s []byte) bool {
	if len(s) == 0 {
		return false
	}

	for _, r := range s {
		if 'a' <= r && r <= 'z' {
			continue
		}
		if 'A' <= r && r <= 'Z' {
			continue
		}
		if '0' <= r && r <= '9' {
			continue
		}

		switch r {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			continue
		default:
			return false
		}
	}

	return true
}

type Headers struct {
	headers map[string]string
}

var rn = []byte("\r\n")

func NewHeaders() *Headers {
	return &Headers{
		headers: map[string]string{},
	}
}

func (h *Headers) Get(name string) (string, bool) {

	str, ok := h.headers[strings.ToLower(name)]
	return str, ok
}

func (h *Headers) Set(name string, value string) {
	v, ok := h.Get(name)
	name = strings.ToLower(name)

	if !ok {
		h.headers[name] = value
		return
	} else {

		h.headers[name] = fmt.Sprintf("%s,%s", v, value)
	}
}

func (h *Headers) ForEach(cb func(n, v string)) {
	for k, v := range h.headers {
		cb(k, v)
	}
}

func parseHeader(fieldLine []byte) (string, string, error) {
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("malformed header")
	}
	name := parts[0]
	value := bytes.TrimSpace(parts[1])
	if !bytes.Equal(name, bytes.TrimSpace(name)) {
		return "", "", fmt.Errorf("malformed key in header")
	}

	return string(name), string(value), nil

}

func (h *Headers) Parse(data []byte) (int, bool, error) {

	read := 0
	isDone := false

	for {
		idx := bytes.Index(data[read:], rn)

		if idx == -1 {
			break
		}
		// Empty header
		if idx == 0 {
			isDone = true
			read += len(rn)
			break
		}

		name, value, err := parseHeader(data[read : read+idx])

		if err != nil {
			return 0, false, err
		}

		if !isToken([]byte(name)) {
			return 0, false, fmt.Errorf("malformed header name")
		}

		read += idx + len(rn)

		h.Set(name, value)

	}

	return read, isDone, nil
}
