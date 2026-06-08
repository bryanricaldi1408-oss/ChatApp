package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

var (
	clients      = make(map[net.Conn]User)
	clientsMutex sync.Mutex
)

func Broadcast(senderConn net.Conn, senderName string, message string) {
	clientsMutex.Lock()
	for conn := range clients {
		if conn != senderConn {
			fmt.Fprintf(conn, "\n[%s]: %s", senderName, message)
		}
	}
	clientsMutex.Unlock()
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// baca username pertama kali
	username, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Failed to read username")
		return
	}

	username = strings.TrimSpace(username)
	clientsMutex.Lock()
	clients[conn] = User{
		conn:     conn,
		username: username,
		room:     nil,
	}
	clientsMutex.Unlock()

	fmt.Printf("%s joined\n", username)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("%s disconnected\n", username)

			clientsMutex.Lock()
			delete(clients, conn)
			clientsMutex.Unlock()
			return
		}

		fmt.Printf("%s", message)

		Broadcast(conn, username, message)
	}
}
func main() {
	ln, err := net.Listen("tcp", ":9090")
	if err != nil {
		fmt.Fprint(os.Stderr, "Failed to listen! ", err)
		os.Exit(1)
	} else {
		fmt.Println("Listening...")
	}
	for {
		conn, err := ln.Accept()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to accept connection:%v \n", err)
			continue
		} else {
			fmt.Println("New connection accepted!")
		}
		go handleConnection(conn)
	}

}
