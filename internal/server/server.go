package server

import (
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
type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	closed  atomic.Bool
	handler Handler
}

func runConnections(s *Server, conn io.ReadWriteCloser) {
	defer conn.Close()

	responseWriter := response.NewWriter(conn)

	req, err := request.RequestFromReader(conn)

	if err != nil {
		responseWriter.WriteStatusLine(response.StatusBarRequest)
		responseWriter.WriteHeaders(response.GetDefaultHeaders(0))
		return
	}
	s.handler(responseWriter, req)

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
