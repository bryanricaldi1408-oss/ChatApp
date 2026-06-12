package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

type Client struct {
	conn net.Conn
	name string
}

type Server struct {
	clients   map[*Client]bool
	broadcast chan Message
	mu        sync.Mutex
}
type Message struct {
	sender  *Client
	content string
}

func handleConnection(conn net.Conn, chatServer *Server) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	usernameInput, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Gagal membaca username dari client")
		return
	}

	username := strings.TrimSpace(usernameInput)

	newClient := Client{
		conn: conn,
		name: username,
	}

	chatServer.mu.Lock()
	chatServer.clients[&newClient] = true
	for client := range chatServer.clients {
		fmt.Fprintf(client.conn, "Klien %s terhubung ke server\n", newClient.name)
	}
	chatServer.mu.Unlock()

	for {
		message, err := reader.ReadString('\n')
		chatServer.mu.Lock()
		if err != nil {
			fmt.Printf("Kilen %s terputus\n", newClient.name)
			delete(chatServer.clients, &newClient)
			break
		}
		chatServer.mu.Unlock()
		formattedMsg := fmt.Sprintf("\n[%s]: %s", newClient.name, message)
		msgObj := Message{
			sender:  &newClient,
			content: formattedMsg,
		}

		chatServer.broadcast <- msgObj
	}
}

func handleMessage(chatServer *Server) {
	for {
		msg := <-chatServer.broadcast
		chatServer.mu.Lock()
		for client := range chatServer.clients {
			if client == msg.sender {
				continue
			}
			_, err := fmt.Fprintf(client.conn, "%s\n", msg.content)
			if err != nil {
				client.conn.Close()
				delete(chatServer.clients, client)
			}
		}
		chatServer.mu.Unlock()
	}
}
func main() {
	chatServer := &Server{
		clients:   make(map[*Client]bool),
		broadcast: make(chan Message),
	}
	port := ":9090"
	ln, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen!\n")
		os.Exit(1)
	}
	defer ln.Close()
	fmt.Printf("Listening at %s\n", port)

	go handleMessage(chatServer)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to accept connection\n")
			continue
		}
		fmt.Println("New connection accepted!")

		go handleConnection(conn, chatServer)

	}
}
