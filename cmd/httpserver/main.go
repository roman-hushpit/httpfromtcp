package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	headers2 "github.com/roman-hushpit/learn-http-protocol/internal/headers"
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
			headers.Set("Trailer", "X-Content-SHA256, X-Content-Length")
			w.WriteHeaders(headers)

			buf := make([]byte, 1024)
			totalBody := make([]byte, 0, 1024)
			for {
				read, err := binResponse.Body.Read(buf)
				if read > 0 {
					totalBody = append(totalBody, buf[:read]...)
					w.WriteChunkedBody(buf[:read])
				}
				if err != nil {
					if err == io.EOF {
						break
					}
					break
				}
			}
			if headers["Trailer"] != "" {
				newHeaders := headers2.NewHeaders()
				sum256 := sha256.Sum256(totalBody)
				newHeaders.Set("X-Content-SHA256", fmt.Sprintf("%x", sum256))
				newHeaders.Set("X-Content-Length", fmt.Sprintf("%d", len(totalBody)))
				w.WriteTrailers(newHeaders)
			} else {
				w.WriteChunkedBodyDone()
			}
			return
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
