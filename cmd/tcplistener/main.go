package main

import (
	"fmt"
	"log"
	"net"

	"github.com/Danjfreire/httpfromtcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:42069")
	if err != nil {
		log.Fatal("Failed to listen")
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Failed to accept connection")
		}

		fmt.Println("Connection started")
		request, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("Failed to read request from reader")
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %v\n", request.RequestLine.Method)
		fmt.Printf("- Target: %v\n", request.RequestLine.RequestTarget)
		fmt.Printf("- Version: %v\n", request.RequestLine.HttpVersion)

		fmt.Printf("Headers:\n")
		for k, v := range request.Headers {
			fmt.Printf("- %v: %v\n", k, v)
		}
		fmt.Printf("Body:")
		fmt.Printf("- %v\n", string(request.Body))

	}

}
