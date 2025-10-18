// Package request
package request

import (
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/Kartik-Saini77/http_server/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        string
	state       parserState
}

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

type parserState string

const (
	StateInit    parserState = "init"
	StateHeaders parserState = "headers"
	StateBody    parserState = "body"
	StateDone    parserState = "done"
	StateError   parserState = "error"
)

func getInt(headers headers.Headers, name string, defaultValue int) int {
	valueStr, exists := headers.Get(name)
	if !exists {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func (r *Request) hasBody() bool {
	length := getInt(r.Headers, "content-length", 0)
	return length > 0
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
		Body:    "",
	}

	buf := make([]byte, 1024)
	bufLen := 0
	for request.state != StateDone && request.state != StateError {
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			return nil, err
		}

		bufLen += n
		readN, err := request.Parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}

	return request, nil
}

func (r *Request) Parse(data []byte) (int, error) {
	read := 0

outer:
	for {
		currentData := data[read:]
		if len(currentData) == 0 {
			break outer
		}

		switch r.state {
		case StateInit:
			rl, n, err := ParseRequestLine(string(currentData))
			if err != nil {
				r.state = StateError
				return 0, err
			}

			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n
			r.state = StateHeaders

		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}

			if n == 0 {
				break outer
			}

			read += n

			if done && r.hasBody() {
				r.state = StateBody
			} else if done {
				r.state = StateDone
			}

		case StateBody:
			length := getInt(r.Headers, "content-length", 0)
			if length == 0 {
				panic("chunked not implemented")
			}
			
			remaining := min(length - len(r.Body), len(currentData))
			r.Body += string(currentData[:remaining])
			read += remaining

			if len(r.Body) == length {
				r.state = StateDone
			}

		case StateDone:
			break outer

		case StateError:
			return 0, errors.New("request in error state")
		}
	}

	return read, nil
}

func ParseRequestLine(line string) (*RequestLine, int, error) {
	idx := strings.Index(line, "\r\n")
	if idx == -1 {
		return nil, 0, nil
	}

	line = line[:idx]
	requestLine := strings.Split(line, " ")
	if len(requestLine) != 3 {
		return nil, 0, errors.New("bad start line")
	}

	for _, r := range requestLine[0] {
		if r < 'A' || r > 'Z' {
			return nil, 0, errors.New("invalid request method")
		}
	}
	if requestLine[2] != "HTTP/1.1" {
		return nil, 0, errors.New("invalid HTTP version")
	}

	return &RequestLine{
		HTTPVersion:   requestLine[2][5:],
		RequestTarget: requestLine[1],
		Method:        requestLine[0],
	}, idx + 2, nil
}
