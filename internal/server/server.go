package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/Danjfreire/httpfromtcp/internal/request"
	"github.com/Danjfreire/httpfromtcp/internal/response"
)

type HandlerError struct {
	Code    response.StatusCode
	Message string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func (he *HandlerError) Write(w io.Writer) {
	response.WriteStatusLine(w, he.Code)
	body := []byte(he.Message)
	headers := response.GetDefaultHeaders(len(body))
	response.WriteHeaders(w, headers)
	w.Write(body)
}

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

	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Fatal("Failed to extract request from connection")
	}

	var bodyBuf bytes.Buffer
	handlerErr := s.handler(&bodyBuf, req)
	if handlerErr != nil {
		handlerErr.Write(conn)
		return
	}

	headers := response.GetDefaultHeaders(bodyBuf.Len())

	err = response.WriteStatusLine(conn, response.StatusOk)
	if err != nil {
		fmt.Println("failed to write status line")
	}

	err = response.WriteHeaders(conn, headers)
	if err != nil {
		fmt.Printf("failed to write response headers: %v\n", err)
	}

	conn.Write(bodyBuf.Bytes())
}
