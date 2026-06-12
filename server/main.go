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
	room string
}

type Room struct {
	Name    string
	Owner   *Client
	Members map[*Client]bool
}

type Server struct {
	clients   map[*Client]bool
	rooms     map[string]*Room
	broadcast chan Message
	mutex     sync.RWMutex
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

	chatServer.mutex.Lock()
	chatServer.clients[&newClient] = true
	chatServer.mutex.Unlock()
	fmt.Printf("Klien %s telah terhubung\n", newClient.name)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Klien %s terputus\n", newClient.name)
			chatServer.mutex.Lock()
			delete(chatServer.clients, &newClient)
			chatServer.mutex.Unlock()
			break
		}

		message = strings.TrimSpace(message)
		if strings.HasPrefix(message, "/create ") {
			roomName := strings.TrimSpace(
				strings.TrimPrefix(message, "/create "),
			)
			createRoom(&newClient, roomName, chatServer)
			continue
		}

		if strings.HasPrefix(message, "/deleteroom ") {
			roomName := strings.TrimSpace(
				strings.TrimPrefix(message,
					"/deleteroom "),
			)

			deleteRoom(&newClient, roomName, chatServer)
			continue
		}

		if message == "/rooms" {
			listRooms(&newClient, chatServer)
			continue
		}

		if strings.HasPrefix(message, "/join ") {
			roomName := strings.TrimSpace(
				strings.TrimPrefix(message, "/join "),
			)
			joinRoom(&newClient, roomName, chatServer)
			continue
		}

		if message == "/leave" {
			leaveRoom(&newClient, chatServer)
			continue
		}

		if message == "/who" {
			whoInRoom(&newClient, chatServer)
			continue
		}

		if newClient.room == "" {
			fmt.Fprintf(conn,
				"[SERVER] Anda harus masuk room terlebih dahulu. Gunakan /join <nama_room>\n")
			continue
		}

		formattedMsg := fmt.Sprintf("[%s]: %s", newClient.name, message)
		msgObj := Message{
			sender:  &newClient,
			content: formattedMsg,
		}

		//ini buat debug
		//fmt.Printf("SERVER RECEIVED: [%s] %s\n", newClient.name, message)

		chatServer.broadcast <- msgObj
	}
}

func handleMessage(chatServer *Server) {
	for {
		msg := <-chatServer.broadcast
		//ini buat debug
		//fmt.Printf("BROADCAST CONTENT = %#v\n", msg.content)
		if msg.sender.room == "" {
			continue
		}
		room := chatServer.rooms[msg.sender.room]
		for client := range room.Members {
			if client == msg.sender {
				continue
			}
			_, err := fmt.Fprintf(client.conn, "%s\n", msg.content)
			if err != nil {
				client.conn.Close()
				delete(chatServer.clients, client)
			}
		}
	}
}

func joinRoom(client *Client, roomName string, server *Server) {

	server.mutex.Lock()
	defer server.mutex.Unlock()

	room, exists := server.rooms[roomName]

	if !exists {
		fmt.Fprintf(client.conn,
			"Room tidak ditemukan\n")
		return
	}

	if client.room != "" {

		oldRoom := server.rooms[client.room]

		delete(oldRoom.Members, client)
	}

	room.Members[client] = true
	client.room = roomName

	fmt.Fprintf(client.conn,
		"Berhasil masuk room %s\n",
		roomName)
}

func createRoom(client *Client, roomName string, server *Server) {

	server.mutex.Lock()
	defer server.mutex.Unlock()

	if _, exists := server.rooms[roomName]; exists {
		fmt.Fprintf(client.conn,
			"Room %s sudah ada\n",
			roomName)
		return
	}

	server.rooms[roomName] = &Room{
		Name:    roomName,
		Owner:   client,
		Members: make(map[*Client]bool),
	}

	fmt.Fprintf(client.conn,
		"Room %s berhasil dibuat\n",
		roomName)
}

func deleteRoom(client *Client, roomName string, server *Server) {

	server.mutex.Lock()
	defer server.mutex.Unlock()

	room, exists := server.rooms[roomName]

	if !exists {
		fmt.Fprintf(client.conn,
			"Room tidak ditemukan\n")
		return
	}

	if room.Owner != client {
		fmt.Fprintf(client.conn,
			"Anda bukan pemilik room %s\n",
			roomName)
		return
	}

	for member := range room.Members {
		member.room = ""
		fmt.Fprintf(member.conn,
			"[SERVER] Room %s telah dihapus\n",
			roomName)
	}

	delete(server.rooms, roomName)
	fmt.Fprintf(client.conn,
		"Room %s berhasil dihapus\n",
		roomName)
}

func listRooms(client *Client, server *Server) {

	server.mutex.RLock()
	defer server.mutex.RUnlock()

	if len(server.rooms) == 0 {
		fmt.Fprintf(client.conn,
			"Tidak ada room. Untuk membuat room, gunakan /create <nama_room>\n")
		return
	}

	fmt.Fprintf(client.conn,
		"Daftar room:\n")

	for _, room := range server.rooms {

		fmt.Fprintf(client.conn,
			"- %s | owner=%s | users=%d\n",
			room.Name,
			room.Owner.name,
			len(room.Members))
	}
}

func leaveRoom(client *Client, server *Server) {

	if client.room == "" {
		fmt.Fprintf(client.conn,
			"Anda tidak berada di room manapun\n")
		return
	}

	room := server.rooms[client.room]
	delete(room.Members, client)

	oldRoom := client.room
	client.room = ""

	fmt.Fprintf(client.conn,
		"Keluar dari room %s\n",
		oldRoom)
}

func whoInRoom(client *Client, server *Server) {

	if client.room == "" {
		fmt.Fprintf(client.conn,
			"Anda belum masuk room\n")
		return
	}

	room := server.rooms[client.room]

	fmt.Fprintf(client.conn,
		"Anggota room %s:\n",
		room.Name)

	for member := range room.Members {

		label := ""

		if member == room.Owner {
			label = " (OWNER)"
		}

		fmt.Fprintf(client.conn,
			"- %s%s\n",
			member.name,
			label)
	}
}

func main() {
	chatServer := &Server{
		clients:   make(map[*Client]bool),
		rooms:     make(map[string]*Room),
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
