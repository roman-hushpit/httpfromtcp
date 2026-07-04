package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/roman-hushpit/learn-http-protocol/internal/request"
	"github.com/roman-hushpit/learn-http-protocol/internal/response"
	"github.com/roman-hushpit/learn-http-protocol/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, func(w io.Writer, req *request.Request) *server.HandlerError {
		if req.RequestLine.RequestTarget == "/yourproblem" {
			return &server.HandlerError{
				Message:    "Your problem is not my problem\n",
				StatusCode: response.BAD_REQUEST,
			}
		}
		if req.RequestLine.RequestTarget == "/myproblem" {
			return &server.HandlerError{
				Message:    "Woopsie, my bad\n",
				StatusCode: response.INTERNAL_SERVER_ERROR,
			}
		}
		w.Write([]byte("All good, frfr\n"))
		return nil
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
