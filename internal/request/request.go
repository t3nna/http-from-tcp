package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
}

var SEPARATOR = "\r\n"
var ERROR_BAD_START_LINE = fmt.Errorf("malformed request-line")
var ERROR_UNSUPPORTED_HPPT_VERSION = fmt.Errorf("unsupported http version")

func (rl *RequestLine) ValidHttp() bool {
	return rl.HttpVersion == "1.1"
}
func (rl *RequestLine) ValidMethod() bool {
	return rl.Method == strings.ToUpper(rl.Method)
}

func parseRequestLine(req string) (*RequestLine, string, error) {
	idx := strings.Index(req, SEPARATOR)
	startLine := req[:idx]
	restOfMsg := req[idx+len(SEPARATOR):]

	parts := strings.Split(startLine, " ")
	if len(parts) != 3 {
		return nil, "", ERROR_BAD_START_LINE
	}

	httpParts := strings.Split(parts[2], "/")
	if len(httpParts) != 2 {
		return nil, "", ERROR_UNSUPPORTED_HPPT_VERSION
	}

	fmt.Println(parts)
	rl := &RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HttpVersion:   httpParts[1],
	}
	if !rl.ValidHttp() {
		return nil, "", ERROR_UNSUPPORTED_HPPT_VERSION
	}

	if !rl.ValidMethod() {
		return nil, "", ERROR_BAD_START_LINE
	}

	return rl, restOfMsg, nil

}

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)

	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("unable to io.ReadAll"), err)
	}

	str := string(data)

	rl, _, err := parseRequestLine(str)
	if err != nil {
		return nil, err
	}

	return &Request{RequestLine: *rl}, err
}
