// Package server
package server

import (
	"fmt"
	"log"
	"net"

	"github.com/Kartik-Saini77/http_server/internal/request"
	"github.com/Kartik-Saini77/http_server/internal/response"
)

type Server struct {
	listener net.Listener
	handler  Handler
	closed   bool
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w *response.Writer, req *request.Request)

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	server := &Server{
		listener: listener,
		handler:  handler,
		closed:   false,
	}
	go server.listen()

	return server, nil
}

func (s *Server) Close() error {
	s.closed = true
	err := s.listener.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) listen() {
	for {
		client, err := s.listener.Accept()
		if s.closed {
			return
		}
		if err != nil {
			log.Println("Error: ", err)
		}
		log.Println("Client connected: ", client.RemoteAddr().String())

		go s.handle(client)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	responseWriter := response.NewWriter(conn)
	r, err := request.RequestFromReader(conn)
	if err != nil {
		headers := response.GetDefaultHeaders(0)
		responseWriter.WriteStatusLine(response.BadRequest)
		responseWriter.WriteHeaders(headers)
		return
	}

	s.handler(responseWriter, r)
}
