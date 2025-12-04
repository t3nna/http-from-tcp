package server

import (
	"fmt"
	"io"
	"net"
	"sync/atomic"
)

type Server struct {
	closed atomic.Bool
}

func runConnections(s *Server, conn io.ReadWriteCloser) {
	out := []byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello World!")
	conn.Write(out)
	conn.Close()
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

func Serve(port uint16) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{closed: atomic.Bool{}}
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
