package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)


type Client struct{
	conn net.Conn
	name string
}

type Server struct{
	clients map[*Client]bool	
	broadcast chan string
}

func handleConnection(conn net.Conn, chatServer *Server){
	defer conn.Close()	

	reader := bufio.NewReader(conn)
	
	fmt.Fprintf(conn, "Silahkan masukkan username Anda: ")

	usernameInput, err := reader.ReadString('\n')
	if err != nil{
		fmt.Println("Gagal membaca username dari client")
		return
	}

	username := strings.TrimSpace(usernameInput)

	newClient := Client{
		conn : conn,
		name : username,
	}

	chatServer.clients[&newClient]= true
}

func main(){
	chatServer := &Server{
		clients: make(map[*Client]bool),
		broadcast: make(chan string),
	}
	port := ":9090"
	ln, err := net.Listen("tcp", port)
	if err != nil{
		fmt.Fprintf(os.Stderr, "Failed to listen!\n")
		os.Exit(1)
	}
	defer ln.Close()	
	fmt.Printf("Listening at %s\n", port)

	for{
		conn, err := ln.Accept()
		if err != nil{
			fmt.Fprintf(os.Stderr, "Failed to accept connection\n")
			continue
		}
		fmt.Println("New connection accepted!")
		
		go handleConnection(conn, chatServer)
				
	}
}
