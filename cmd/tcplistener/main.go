package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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
		lines := getLinesChannel(conn)
		for line := range lines {
			fmt.Printf("read: %s\n", line)
		}
	}

}

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)
	go func() {
		defer f.Close()
		defer fmt.Println("Channel Closed")
		defer close(lines)

		data := make([]byte, 8)
		offset := 0
		read := -1
		currentLine := ""

		for read != 0 {
			read, err := f.Read(data)
			offset += read
			currentLine += string(data[:read])

			if err == io.EOF {
				break
			}

			parts := strings.Split(currentLine, "\n")
			if len(parts) > 1 {
				lines <- parts[0]
				currentLine = strings.Join(parts[1:], "")
			}
		}

		if len(currentLine) != 0 {
			lines <- currentLine
		}

	}()

	return lines
}
