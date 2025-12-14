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

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()

	headers["Content-Length"] = fmt.Sprint(contentLen)
	headers["Connection"] = "close"
	headers["Content-Type"] = "text/plain"

	return headers
}
