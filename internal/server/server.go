package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/Danjfreire/httpfromtcp/internal/request"
	"github.com/Danjfreire/httpfromtcp/internal/response"
)

type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
	handler  Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		listener: listener,
		handler:  handler,
	}

	go s.listen()
	return s, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	if s.listener != nil {
		return s.listener.Close()
	}

	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("Error accepting connection: %v\n", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	res := &response.Writer{
		Writer: conn,
	}

	req, err := request.RequestFromReader(conn)
	if err != nil {
		res.WriteStatusLine(response.StatusBadRequest)
		res.WriteHeaders(response.GetDefaultHeaders(0))
		return
	}
	s.handler(res, req)
}
