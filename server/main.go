package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	server := NewServer()

	port := ":9090"
	ln, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen!\n")
		os.Exit(1)
	}
	defer ln.Close()
	fmt.Printf("Listening at %s\n", port)

	go server.HandleMessage()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to accept connection\n")
			continue
		}
		fmt.Println("New connection accepted!")

		go HandleConnection(conn, server)
	}
}
