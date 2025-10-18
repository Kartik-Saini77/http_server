package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/Kartik-Saini77/http_server/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Error starting tcp server: ", err)
	}
	defer listener.Close()
	log.Println("Server started on port 8080")

	for {
		client, err := listener.Accept()
		if err != nil {
			log.Println("Error: ", err)
		}
		log.Println("Client connected: ", client.RemoteAddr().String())

		r, err := request.RequestFromReader(client)
		if err != nil {
			log.Println("Error: ", err)
		}

		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %s\n", r.RequestLine.Method)
		fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", r.RequestLine.HTTPVersion)
		
		fmt.Printf("\nHeaders:\n")
		for header, value := range r.Headers {
			fmt.Printf("- %s: %s\n", header, value)
		}
		
		fmt.Printf("\nBody:\n")
		fmt.Printf("%s\n", r.Body)
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)
		defer f.Close()
		str := ""
		for {
			data := make([]byte, 8)
			n, err := f.Read(data)
			if err != nil {
				break
			}

			data = data[:n]
			if i := bytes.IndexByte(data, '\n'); i != -1 {
				str += string(data[:i])
				data = data[i+1:]
				ch <- str
				str = ""
			}
			str += string(data)
		}
		if len(str) != 0 {
			ch <- str
		}
	}()

	return ch
}
