package response

import (
	"fmt"
	"io"

	"github.com/Danjfreire/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusOk StatusCode = iota
	StatusBadRequest
	StatusInternalServerError
)

type Writer struct {
	Writer io.Writer
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	var err error
	switch statusCode {
	case StatusOk:
		_, err = fmt.Fprint(w.Writer, "HTTP/1.1 200 OK\r\n")
	case StatusBadRequest:
		_, err = fmt.Fprint(w.Writer, "HTTP/1.1 400 Bad Request\r\n")
	case StatusInternalServerError:
		_, err = fmt.Fprint(w.Writer, "HTTP/1.1 500 Internal Server Error\r\n")
	default:
		_, err = fmt.Fprint(w.Writer, "")
	}

	if err != nil {
		return err
	}

	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	for k, v := range headers {
		_, err := fmt.Fprintf(w.Writer, "%v: %v\r\n", k, v)
		if err != nil {
			return err
		}
	}

	fmt.Fprint(w.Writer, "\r\n")
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	return w.Writer.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	chunkSize := len(p)

	nTotal := 0
	n, err := fmt.Fprintf(w.Writer, "%x\r\n", chunkSize)
	if err != nil {
		return nTotal, err
	}
	nTotal += n

	n, err = w.Writer.Write(p)
	if err != nil {
		return nTotal, err
	}
	nTotal += n

	n, err = w.Writer.Write([]byte("\r\n"))
	if err != nil {
		return nTotal, err
	}
	nTotal += n
	return nTotal, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	n, err := w.Writer.Write([]byte("0\r\n"))
	if err != nil {
		return n, err
	}
	return n, nil
}

func (w *Writer) WriteTrailers(headers headers.Headers) error {
	for k, v := range headers {
		_, err := fmt.Fprintf(w.Writer, "%v: %v\r\n", k, v)
		if err != nil {
			return err
		}
	}

	fmt.Fprint(w.Writer, "\r\n")
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()

	headers.Set("Content-Length", fmt.Sprint(contentLen))
	headers.Set("Connection", "close")
	headers.Set("Content-Type", "text/plain")

	return headers
}
