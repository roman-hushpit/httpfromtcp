package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/roman-hushpit/learn-http-protocol/internal/request"
	"github.com/roman-hushpit/learn-http-protocol/internal/response"
)

type Server struct {
	Port     int
	Listener net.Listener
	Closed   atomic.Bool
	Handler  Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	server := &Server{
		Port:     port,
		Listener: listener,
		Handler:  handler,
	}

	go server.listen()

	return server, nil
}

func (s *Server) Close() error {
	s.Closed.Store(true)
	err := s.Listener.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) listen() {
	for {
		if s.Closed.Load() {
			break
		}
		conn, err := s.Listener.Accept()
		if err != nil {
			log.Printf("error: %s\n", err.Error())
			continue
		}
		fmt.Println("Accepted connection from", conn.RemoteAddr())
		go s.handle(conn)

	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	parsedRequest, err := request.RequestFromReader(conn)
	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.BAD_REQUEST,
			Message:    err.Error(),
		}
		hErr.WriteError(conn)
		return
	}
	newWriter := response.NewWriter(conn)
	s.Handler(newWriter, parsedRequest)
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (he *HandlerError) WriteError(out io.Writer) {
	err := response.WriteStatusLine(out, he.StatusCode)
	if err != nil {
		return
	}
	err = response.WriteHeaders(out, response.GetDefaultHeaders(len([]byte(he.Message))))
	if err != nil {
		return
	}
	_, err = out.Write([]byte(he.Message))
	if err != nil {
		return
	}
}

type Handler func(w *response.Writer, req *request.Request)
