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
	server.mutex.Lock()
	for existingClient := range server.clients {
		if client.name == existingClient.name {
			server.mutex.Unlock()
			fmt.Fprintf(client.conn, "Username %s sudah dipakai\n", existingClient.name)
			return
		}
	}
	server.mutex.Unlock()

	server.AddClient(client)

	server.mutex.Lock()
	for allClient := range server.clients {
		fmt.Fprintf(allClient.conn, "[SERVER] %s berhasil bergabung ke server\n", client.name)
	}
	server.mutex.Unlock()
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Klien %s terputus\n", client.name)
			server.RemoveClient(client)
			break
		}

		message = strings.TrimSpace(message)

		switch {
		case message == "/help":
			helpMsg := `
Daftar perintah yang tersedia:
/create <nama_room>    - Membuat room baru
/deleteroom <nama_room> - Menghapus room (hanya pemilik)
/rooms                 - Menampilkan daftar semua room yang tersedia
/join <nama_room>      - Bergabung ke dalam room
/leave                 - Keluar dari room saat ini
/who                   - Menampilkan daftar pengguna di room saat ini
/help                  - Menampilkan daftar perintah ini
/quit                  - Keluar dari server
`
			fmt.Fprintf(conn, "%s\n", helpMsg)
			continue

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
				"[SERVER] Anda harus masuk room terlebih dahulu. Gunakan /join <nama_room> atau gunakan perintah /help untuk melihat semua command\n")
			continue
		}

		formattedMsg := fmt.Sprintf("[%s]: %s", client.name, message)

		// Mengirim pesan ke channel broadcast menggunakan operator <- .
		// Pesan akan diproses lebih lanjut oleh goroutine server.HandleMessage().
		// Referensi: https://go.dev/tour/concurrency/2
		server.broadcast <- Message{
			sender:  client,
			content: formattedMsg,
		}
	}
}
