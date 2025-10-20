// Package headers
package headers

import (
	"bytes"
	"errors"
	"strings"
	"unicode"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Get(key string) (value string, ok bool) {
	header, ok := h[strings.ToLower(key)]
	return header, ok 
}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)
	if val, ok := h[key]; ok {
		h[key] = val + ", " + value
	} else {
		h[key] = value
	}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return 0, false, nil
	}

	if idx == 0 {
		return 2, true, nil
	}

	fieldLine := string(data[:idx])

	colonIdx := strings.Index(fieldLine, ":")
	if colonIdx == -1 {
		return 0, false, errors.New("invalid field-line")
	}

	key := fieldLine[:colonIdx]
	value := strings.TrimSpace(fieldLine[colonIdx+1:])

	if strings.TrimSpace(key) != key {
        return 0, false, errors.New("invalid field-line")
    }
	
	for _, r := range key {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("!#$%&'*+-.^_`|~", r) {
			continue
		}
		return 0, false, errors.New("invalid header key")
	}

	h.Set(key, value)

	return idx + 2, false, nil
}
