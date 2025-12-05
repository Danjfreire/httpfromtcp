package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatalf("Failed to resolve UDP addr : %v", err)
	}

	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatalf("Failed to Dial UDP: %v", err)
	}

	defer udpConn.Close()

	reader := bufio.NewReader(os.Stdin)
	// data := []byte{}
	for {
		fmt.Printf(">")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Failed to read data\n")
		}

		_, err = udpConn.Write([]byte(line))
		if err != nil {
			log.Fatalf("Failed to write to udp connection: %v\n", err)
		}
	}
}
