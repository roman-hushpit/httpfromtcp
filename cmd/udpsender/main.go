package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatalf("error resolving UDP address: %s\n", err.Error())
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatalf("error connecting to UDP: %s\n", err.Error())
	}
	defer conn.Close()
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">")
		readString, err := reader.ReadString('\n')
		if err != nil {
			log.Println("error reading input: " + err.Error())
			return
		}
		write, err := conn.Write([]byte(readString))
		if err != nil {
			log.Println("error writing input: " + err.Error())
			return
		}
		log.Println(write)

	}
}
