package server

import (
	"bytes"
	"fmt"
	"github.com/t3nna/http-from-tcp/internal/request"
	"github.com/t3nna/http-from-tcp/internal/response"
	"io"
	"net"
	"sync/atomic"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}
type Handler func(w io.Writer, req *request.Request) *HandlerError

type Server struct {
	closed  atomic.Bool
	handler Handler
}

func runConnections(s *Server, conn io.ReadWriteCloser) {
	defer conn.Close()

	headers := response.GetDefaultHeaders(0)
	fmt.Println(headers)

	req, err := request.RequestFromReader(conn)
	fmt.Println(req, err)
	if err != nil {
		response.WriteStatusLine(conn, response.StatusBarRequest)
		response.WriteHeaders(conn, headers)
		return
	}
	writer := bytes.NewBuffer([]byte{})
	handlerError := s.handler(writer, req)

	var body []byte = nil
	var status response.StatusCode = response.StatusOK
	if handlerError != nil {
		status = handlerError.StatusCode
		body = []byte(handlerError.Message)
	} else {
		body = writer.Bytes()
	}

	headers.Replace("Content-Length", fmt.Sprintf("%d", len(body)))

	response.WriteStatusLine(conn, status)
	response.WriteHeaders(conn, headers)
	conn.Write(body)
}

func runServer(s *Server, listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if s.closed.Load() {
			return
		}

		if err != nil {
			return
		}

		go runConnections(s, conn)

	}

}

func Serve(port uint16, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		closed:  atomic.Bool{},
		handler: handler,
	}
	go runServer(s, listener)

	return s, nil
}

func (s *Server) Close() error {

	s.closed.Store(false)
	return nil
}

func (s *Server) listen() {

}
func (s *Server) handle(conn net.Conn) {

}
