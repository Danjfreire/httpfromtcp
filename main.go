package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("messages.txt")
	if err != nil {
		log.Fatalf("Failed to open file : %v", err)
	}

	lines := getLinesChannel(file)

	for line := range lines {
		fmt.Printf("read: %s\n", line)
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)
	go func() {
		defer f.Close()
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
