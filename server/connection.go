package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// membaca username, lalu memproses command/pesan secara terus-menerus.
func HandleConnection(conn net.Conn, server *Server) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	usernameInput, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Gagal membaca username dari client")
		return
	}

	username := strings.TrimSpace(usernameInput)

	client := &Client{
		conn: conn,
		name: username,
	}

	server.AddClient(client)
	fmt.Printf("Klien %s telah terhubung\n", client.name)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Klien %s terputus\n", client.name)
			server.RemoveClient(client)
			break
		}

		message = strings.TrimSpace(message)

		switch {
		case strings.HasPrefix(message, "/create "):
			roomName := strings.TrimSpace(strings.TrimPrefix(message, "/create "))
			server.CreateRoom(client, roomName)
			continue

		case strings.HasPrefix(message, "/deleteroom "):
			roomName := strings.TrimSpace(strings.TrimPrefix(message, "/deleteroom "))
			server.DeleteRoom(client, roomName)
			continue

		case message == "/rooms":
			server.ListRooms(client)
			continue

		case strings.HasPrefix(message, "/join "):
			roomName := strings.TrimSpace(strings.TrimPrefix(message, "/join "))
			server.JoinRoom(client, roomName)
			continue

		case message == "/leave":
			server.LeaveRoom(client)
			continue

		case message == "/who":
			server.WhoInRoom(client)
			continue
		}

		if client.room == "" {
			fmt.Fprintf(conn,
				"[SERVER] Anda harus masuk room terlebih dahulu. Gunakan /join <nama_room>\n")
			continue
		}

		formattedMsg := fmt.Sprintf("[%s]: %s", client.name, message)
		server.broadcast <- Message{
			sender:  client,
			content: formattedMsg,
		}
	}
}
