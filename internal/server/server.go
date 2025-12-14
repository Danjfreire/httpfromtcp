package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/Danjfreire/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		listener: listener,
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
	headers := response.GetDefaultHeaders(0)

	err := response.WriteStatusLine(conn, response.StatusOk)
	if err != nil {
		fmt.Println("Failed to write status line")
	}

	err = response.WriteHeaders(conn, headers)
	if err != nil {
		fmt.Println("Failed to write response headers")
	}
}
