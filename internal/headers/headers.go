package headers

import (
	"bytes"
	"fmt"
)

type Headers map[string]string

var rn = []byte("\r\n")

func NewHeaders() Headers {
	fmt.Println("call")
	return Headers{}
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

func (h Headers) Parse(data []byte) (int, bool, error) {

	read := 0
	isDone := false

	for {
		fmt.Println("read...")
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
		read += idx + len(rn)

		h[name] = value

	}

	return read, isDone, nil
}
