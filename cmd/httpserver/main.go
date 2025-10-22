package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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
