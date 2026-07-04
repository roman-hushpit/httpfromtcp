package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/roman-hushpit/learn-http-protocol/internal/request"
	"github.com/roman-hushpit/learn-http-protocol/internal/response"
	"github.com/roman-hushpit/learn-http-protocol/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		if req.RequestLine.RequestTarget == "/yourproblem" {
			badRequest := "<html>\n  <head>\n    <title>400 Bad Request</title>\n  </head>\n  <body>\n    <h1>Bad Request</h1>\n    <p>Your request honestly kinda sucked.</p>\n  </body>\n</html>\n"
			w.WriteStatusLine(400)
			headers := response.GetDefaultHeaders(len(badRequest))
			headers.Set("Content-Type", "text/html")
			w.WriteHeaders(headers)
			w.WriteBody([]byte(badRequest))
			return
		}
		if req.RequestLine.RequestTarget == "/myproblem" {
			internalServerError := "<html>\n  <head>\n    <title>500 Internal Server Error</title>\n  </head>\n  <body>\n    <h1>Internal Server Error</h1>\n    <p>Okay, you know what? This one is on me.</p>\n  </body>\n</html>\n"
			w.WriteStatusLine(500)
			headers := response.GetDefaultHeaders(len(internalServerError))
			headers.Set("Content-Type", "text/html")
			w.WriteHeaders(headers)
			w.WriteBody([]byte(internalServerError))
			return
		}

		if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
			path := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
			url := "https://httpbin.org" + path
			binResponse, err := http.Get(url)
			if err != nil {
				return
			}
			w.WriteStatusLine(200)
			headers := response.GetDefaultHeaders(0)
			headers.Set("Content-Type", "text/html")
			headers.Drop("Content-Length")
			headers.Set("Transfer-Encoding", "chunked")
			w.WriteHeaders(headers)

			buf := make([]byte, 1024)
			for {
				read, err := binResponse.Body.Read(buf)
				if read > 0 {
					w.WriteChunkedBody(buf[:read])
				}
				if err != nil {
					if err == io.EOF {
						break
					}
					break
				}
			}
			w.WriteChunkedBodyDone()
		}

		successResponse := "<html>\n  <head>\n    <title>200 OK</title>\n  </head>\n  <body>\n    <h1>Success!</h1>\n    <p>Your request was an absolute banger.</p>\n  </body>\n</html>\n"
		w.WriteStatusLine(200)
		headers := response.GetDefaultHeaders(len(successResponse))
		headers.Set("Content-Type", "text/html")
		w.WriteHeaders(headers)
		w.WriteBody([]byte(successResponse))
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
