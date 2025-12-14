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

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var err error
	switch statusCode {
	case StatusOk:
		_, err = fmt.Fprint(w, "HTTP/1.1 200 OK\r\n")
	case StatusBadRequest:
		_, err = fmt.Fprint(w, "HTTP/1.1 400 Bad Request\r\n")
	case StatusInternalServerError:
		_, err = fmt.Fprint(w, "HTTP/1.1 500 Internal Server Error\r\n")
	default:
		_, err = fmt.Fprint(w, "")
	}

	if err != nil {
		return err
	}

	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()

	headers["Content-Length"] = fmt.Sprint(contentLen)
	headers["Connection"] = "close"
	headers["Content-Type"] = "text/plain"

	return headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for k, v := range headers {
		_, err := fmt.Fprintf(w, "%v: %v\r\n", k, v)
		if err != nil {
			return err
		}
	}

	fmt.Fprint(w, "\r\n")
	return nil
}
