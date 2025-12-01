package request

import (
	"bytes"
	"fmt"
	"github.com/t3nna/http-from-tcp/internal/headers"
	"io"
	"strings"
)

type parserState string

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	state       parserState
}

func newRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
	}
}

var SEPARATOR = []byte("\r\n")
var ERROR_BAD_START_LINE = fmt.Errorf("malformed request-line")
var ERROR_UNSUPPORTED_HPPT_VERSION = fmt.Errorf("unsupported http version")
var bufferSize = 1024

const (
	StateInit          parserState = "init"
	StateDone          parserState = "done"
	StateError         parserState = "error"
	StateHeaderParsing parserState = "headers"
)

func (rl *RequestLine) ValidHttp() bool {
	return rl.HttpVersion == "1.1"
}
func (rl *RequestLine) ValidMethod() bool {
	return rl.Method == strings.ToUpper(rl.Method)
}

func parseRequestLine(req []byte) (*RequestLine, int, error) {
	idx := bytes.Index(req, SEPARATOR)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := req[:idx]
	consumed := idx + len(SEPARATOR)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ERROR_BAD_START_LINE
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ERROR_UNSUPPORTED_HPPT_VERSION
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	if !rl.ValidMethod() {
		return nil, 0, ERROR_BAD_START_LINE
	}
	//if !rl.ValidHttp() {
	//	return nil, 0, ERROR_BAD_START_LINE
	//}

	return rl, consumed, nil

}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

outer:
	for {
		currData := data[read:]

		switch r.state {
		case StateInit:
			rl, consumed, err := parseRequestLine(currData)
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if consumed == 0 {
				break outer
			}
			r.RequestLine = *rl
			read += consumed

			r.state = StateHeaderParsing

		case StateHeaderParsing:
			n, done, err := r.Headers.Parse(currData)

			read += n

			if err != nil {
				r.state = StateError
				return 0, err
			}
			if n == 0 {
				break outer
			}

			//if !done {
			//	break outer
			//}
			if done {
				r.state = StateDone
			}

		case StateDone:
			break outer
		case StateError:
			return 0, ERROR_BAD_START_LINE

		default:
			panic("skill issue programming")
		}

	}

	return read, nil

}

func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	// NOTE: buffer could get overrun...
	buf := make([]byte, bufferSize)
	readToIdx := 0
	for !request.done() {
		if readToIdx >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIdx:])
		if err == io.EOF {
			request.state = StateDone
			break
		}
		if err != nil {
			return nil, err
		}
		readToIdx += n

		readN, err := request.parse(buf[:readToIdx])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:readToIdx])
		readToIdx -= readN
	}

	return request, nil
}
