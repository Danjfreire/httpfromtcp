package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Danjfreire/httpfromtcp/internal/headers"
	"github.com/Danjfreire/httpfromtcp/internal/request"
	"github.com/Danjfreire/httpfromtcp/internal/response"
	"github.com/Danjfreire/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	if req.RequestLine.RequestTarget == "/yourproblem" {
		handler400(w, req)
		return
	}

	if req.RequestLine.RequestTarget == "/myproblem" {
		handler500(w, req)
		return
	}

	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		handlerHttpBin(w, req)
		return
	}

	handler200(w, req)
}

func handler400(w *response.Writer, req *request.Request) {
	w.WriteStatusLine(response.StatusBadRequest)
	res := []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)

	headers := response.GetDefaultHeaders(len(res))
	headers["Content-Type"] = "text/html"
	w.WriteHeaders(headers)
	w.WriteBody(res)
}

func handler500(w *response.Writer, req *request.Request) {
	w.WriteStatusLine(response.StatusInternalServerError)
	res := []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)

	headers := response.GetDefaultHeaders(len(res))
	headers["Content-Type"] = "text/html"
	w.WriteHeaders(headers)
	w.WriteBody(res)
}

func handler200(w *response.Writer, req *request.Request) {
	w.WriteStatusLine(response.StatusOk)
	res := []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)

	headers := response.GetDefaultHeaders(len(res))
	headers["Content-Type"] = "text/html"
	w.WriteHeaders(headers)
	w.WriteBody(res)
}

func handlerHttpBin(w *response.Writer, req *request.Request) {
	target := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	url := "https://httpbin.org/" + target
	fmt.Println("Proxying to", url)
	resp, err := http.Get(url)
	if err != nil {
		handler500(w, req)
		return
	}
	defer resp.Body.Close()

	w.WriteStatusLine(response.StatusOk)
	h := response.GetDefaultHeaders(0)
	h.Override("Transfer-Encoding", "chunked")
	h.Remove("Content-Length")
	h.Set("Trailer", "X-Content-Sha256, X-Content-Length")
	w.WriteHeaders(h)

	const maxChunkSize = 1024
	buffer := make([]byte, maxChunkSize)
	fullBody := []byte{}
	for {
		n, err := resp.Body.Read(buffer)
		fmt.Println("Read", n, "bytes")
		if n > 0 {
			_, err = w.WriteChunkedBody(buffer[:n])
			if err != nil {
				fmt.Println("Error writing chunked body:", err)
				break
			}

			fullBody = append(fullBody, buffer[:n]...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading response body:", err)
			break
		}
	}
	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		fmt.Println("Error writing chunked body done:", err)
	}

	trailers := headers.NewHeaders()
	hash := sha256.Sum256(fullBody)
	trailers.Set("X-Content-Sha256", hex.EncodeToString(hash[:]))
	trailers.Set("X-Content-Length", fmt.Sprintf("%v", len(fullBody)))
	fmt.Printf("hash is %v\n", hex.EncodeToString(hash[:]))
	fmt.Printf("length : %v\n", len(fullBody))
	w.WriteTrailers(trailers)
}
