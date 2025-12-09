package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/Danjfreire/httpfromtcp/internal/headers"
)

type requestState int

const (
	requestStateInitialized = iota
	requestStateDone
	requestStateParsingHeaders
	requestStateParsingBody
)

type Request struct {
	RequestLine    RequestLine
	state          requestState
	Headers        headers.Headers
	Body           []byte
	bodyLengthRead int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const crlf = "\r\n"

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0

	for r.state != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}

		if n == 0 {
			break
		}

		totalBytesParsed += n
	}

	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateDone:
		return 0, errors.New("cannot parse data from a done state")
	case requestStateInitialized:
		n, requestLine, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}

		// we still need to read more
		if n == 0 {
			return 0, nil
		}

		r.RequestLine = *requestLine
		r.state = requestStateParsingHeaders
		return n, nil
	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)

		if err != nil {
			return 0, err
		}

		if n == 0 {
			return 0, nil
		}

		if done {
			r.state = requestStateParsingBody
		}

		return n, nil
	case requestStateParsingBody:
		contentLenStr, ok := r.Headers.Get("Content-Length")
		if !ok {
			// assume that if no content-length header is present, there is no body
			r.state = requestStateDone
			return len(data), nil
		}
		contentLen, err := strconv.Atoi(contentLenStr)
		if err != nil {
			return 0, fmt.Errorf("malformed Content-Length: %s", err)
		}
		r.Body = append(r.Body, data...)
		r.bodyLengthRead += len(data)
		if r.bodyLengthRead > contentLen {
			return 0, fmt.Errorf("Content-Length too large")
		}
		if r.bodyLengthRead == contentLen {
			r.state = requestStateDone
		}
		return len(data), nil
	default:
		return 0, errors.New("unknown state")
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, 8)
	request := &Request{
		state:   requestStateInitialized,
		Headers: headers.NewHeaders(),
		Body:    make([]byte, 0),
	}
	unparsedBytesInBuffer := 0

	for request.state != requestStateDone {
		// if the buffer is full we need to make it bigger so it can hold more data
		if unparsedBytesInBuffer >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		bytesRead, err := reader.Read(buf[unparsedBytesInBuffer:])
		unparsedBytesInBuffer += bytesRead

		bytesParsed, parseErr := request.parse(buf[:unparsedBytesInBuffer])
		if parseErr != nil {
			return nil, parseErr
		}

		// remove the parsed bytes from the buffer
		copy(buf, buf[bytesParsed:])
		unparsedBytesInBuffer -= bytesParsed

		// fmt.Printf("buffer is \"%v\"\n", string(buf))

		// handle EOF here because it is possible to read bytes AND also reach EOF
		if err != nil {
			// if we reached the end of the stream
			if errors.Is(err, io.EOF) {
				request.state = requestStateDone
				break
			}

			return nil, err
		}
	}

	return request, nil
}

func parseRequestLine(data []byte) (int, *RequestLine, error) {
	// if has no \r\n
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, nil, nil
	}

	requestText := string(data[:idx])
	requestLine, err := getRequestLineFromString(requestText)
	if err != nil {
		return 0, nil, err
	}

	return idx + len(crlf), requestLine, nil
}

func getRequestLineFromString(line string) (*RequestLine, error) {
	parts := strings.Split(line, " ")

	if len(parts) != 3 {
		return nil, errors.New("invalid amount of parts in request line")
	}

	method := parts[0]
	// method should contain only uppercase letters
	for _, char := range method {
		validChar := unicode.IsUpper(char) && unicode.IsLetter(char)

		if !validChar {
			return nil, errors.New("invalid method")
		}
	}

	target := parts[1]

	httpVersion := parts[2]

	if httpVersion != "HTTP/1.1" {
		return nil, errors.New("invalid http version, only HTTP/1.1 is supported")
	}

	return &RequestLine{HttpVersion: "1.1", Method: method, RequestTarget: target}, nil
}
