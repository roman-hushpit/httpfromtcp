package server

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/roman-hushpit/learn-http-protocol/internal/response"
)

type Server struct {
	Port     int
	Listener net.Listener
	Closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	server := &Server{
		Port:     port,
		Listener: listener,
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
	err := response.WriteStatusLine(conn, 200)
	if err != nil {
		return
	}
	err = response.WriteHeaders(conn, response.GetDefaultHeaders(0))
	if err != nil {
		return
	}
}
