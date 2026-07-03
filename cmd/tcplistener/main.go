package main

import (
	"fmt"
	"log"
	"net"

	"github.com/roman-hushpit/learn-http-protocol/internal/request"
)

const port = ":42069"

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("error listening for TCP traffic: %s\n", err.Error())
	}
	defer listener.Close()

	fmt.Println("Listening for TCP traffic on", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("error: %s\n", err.Error())
		}
		fmt.Println("Accepted connection from", conn.RemoteAddr())

		request, err := request.RequestFromReader(conn)
		if err != nil {
			return
		}
		fmt.Println(request)

		fmt.Println("Connection to ", conn.RemoteAddr(), "closed")
	}
}
