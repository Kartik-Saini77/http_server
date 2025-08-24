// Package request
package request

import (
	"errors"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
	state       parserState
}

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

type parserState string

const (
	StateInit  parserState = "init"
	StateDone  parserState = "done"
	StateError parserState = "error"
)

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := &Request{
		state: StateInit,
	}

	buf := make([]byte, 1024)
	bufLen := 0
	for request.state != StateDone && request.state != StateError {
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			return nil, err
		}

		bufLen += n
		readN, err := request.parse(buf[:bufLen+n])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}

	return request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

outer:
	for {
		switch r.state {
		case StateInit:
			rl, n, err := ParseRequestLine(string(data[read:]))
			if err != nil {
				r.state = StateError
				return 0, err
			}

			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n
			r.state = StateDone
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

