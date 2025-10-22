package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Kartik-Saini77/http_server/internal/headers"
	"github.com/Kartik-Saini77/http_server/internal/request"
	"github.com/Kartik-Saini77/http_server/internal/response"
	"github.com/Kartik-Saini77/http_server/internal/server"
)

const port = 8080

func main() {
	server, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		h := response.GetDefaultHeaders(0)
		body := respond200()
		status := response.OK

		if req.RequestLine.RequestTarget == "/yourproblem" {
			status = response.BadRequest
			body = respond400()
		} else if req.RequestLine.RequestTarget == "/myproblem" {
			status = response.InternalServerError
			body = respond500()
		} else if req.RequestLine.RequestTarget == "/video" {
			f, err := os.ReadFile("./video.mp4")
			if err != nil {
				log.Println("Error: ", err.Error())
				body = respond500()
				status = response.InternalServerError
			} else {
				h.Replace("Content-type", "video/mp4")
				h.Replace("Content-length", fmt.Sprintf("%d", len(f)))

				w.WriteStatusLine(response.OK)
				w.WriteHeaders(h)
				w.WriteBody(f)
				return
			}
		} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
			req, err := http.Get("https://httpbin.org/" + req.RequestLine.RequestTarget[len("/httpbin/"):])
			if err != nil {
				body = respond500()
				status = response.InternalServerError
			} else {
				w.WriteStatusLine(response.OK)

				h.Delete("Content-length")
				h.Set("Transfer-encoding", "chunked")
				h.Replace("Content-type", "text/plain")
				h.Set("Trailer", "X-Content-SHA256")
				h.Set("Trailer", "X-Content-Length")
				w.WriteHeaders(h)

				var body []byte
				for {
					data := make([]byte, 32)
					n, err := req.Body.Read(data)
					if n > 0 {
						body = append(body, data[:n]...)
						w.WriteChunkedBody(data[:n])
					}
					if err != nil {
						break
					}
				}
				w.WriteChunkedBodyDone()

				hash := sha256.Sum256(body)
				trailers := headers.NewHeaders()
				trailers.Set("X-Content-SHA256", toStr(hash[:]))
				trailers.Set("X-Content-Length", fmt.Sprint(len(body)))
				w.WriteTrailers(trailers)
				return
			}
		}

		w.WriteStatusLine(status)
		h.Replace("Content-length", fmt.Sprintf("%d", len(body)))
		w.WriteHeaders(h)
		w.WriteBody(body)
	})
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

func toStr(bytes []byte) string {
	out := ""
	for _, b := range bytes {
		out += fmt.Sprintf("%02x", b)
	}
	return out
}

func respond400() []byte {
	return []byte(
		`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request resulted in an error.</p>
  </body>
</html>`,
	)
}

func respond500() []byte {
	return []byte(
		`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>The server encountered an internal error or misconfiguration and was unable to complete your request.</p>
  </body>
</html>`,
	)
}

func respond200() []byte {
	return []byte(
		`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was successfully processed.</p>
  </body>
</html>`,
	)
}
