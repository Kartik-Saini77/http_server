// Package response
package response

import (
	"fmt"
	"io"

	"github.com/Kartik-Saini77/http_server/internal/headers"
)

type StatusCode int

type Writer struct {
	writer io.Writer
}

const (
	OK                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

func NewWriter(writer io.Writer) *Writer {
	return &Writer{writer: writer}
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()

	headers.Set("Content-Length", fmt.Sprint(contentLen))
	headers.Set("Connection", "close")
	headers.Set("Content-Type", "text/html")

	return headers
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	statusLine := fmt.Sprintf("HTTP/1.1 %d ", statusCode)

	switch statusCode {
	case OK:
		statusLine += "OK\r\n"
	case BadRequest:
		statusLine += "Bad Request\r\n"
	case InternalServerError:
		statusLine += "Internal Server Error\r\n"
	}

	_, err := w.writer.Write([]byte(statusLine))

	return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	for header, value := range headers {
		_, err := fmt.Fprintf(w.writer, "%s: %s\r\n", header, value)
		if err != nil {
			return err
		}
	}
	w.writer.Write([]byte("\r\n"))
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	return w.writer.Write(p)
}
