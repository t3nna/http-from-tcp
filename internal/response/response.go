package response

import (
	"fmt"
	"github.com/t3nna/http-from-tcp/internal/headers"
	"io"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBarRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

const rn = "\r\n"

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	statusLine := []byte("")
	switch statusCode {
	case StatusOK:
		statusLine = []byte("HTTP/1.1 200 OK")
	case StatusBarRequest:
		statusLine = []byte("HTTP/1.1 400 Bad Request")

	case StatusInternalServerError:
		statusLine = []byte("HTTP/1.1 500 Internal Server Error")

	default:
		return fmt.Errorf("unknow status code")
	}

	statusLine = append(statusLine, []byte(rn)...)

	_, err := w.Write(statusLine)
	return err

}

func GetDefaultHeaders(contentLen int) *headers.Headers {
	h := headers.NewHeaders()
	h.Replace("content-length", fmt.Sprintf("%d", contentLen))
	h.Set("connection", "close")
	h.Set("content-type", "text/plain")
	return h
}
func WriteHeaders(w io.Writer, h *headers.Headers) error {
	var headersLine []byte
	h.ForEach(func(key, value string) {
		headersLine = fmt.Appendf(headersLine, "%s: %s%s", key, value, rn)
	})
	headersLine = fmt.Append(headersLine, rn)
	_, err := w.Write(headersLine)
	return err
}
