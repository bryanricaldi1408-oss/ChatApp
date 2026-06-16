package main

import (
	"fmt"
	"sync"
)

// Server menyimpan state global aplikasi chat: daftar client, daftar room, dan channel broadcast untuk pengiriman pesan.
type Server struct {
	clients   map[*Client]bool
	rooms     map[string]*Room
	broadcast chan Message
	mutex     sync.Mutex
}

// NewServer membuat instance Server baru dengan state kosong.
func NewServer() *Server {
	return &Server{
		clients:   make(map[*Client]bool),
		rooms:     make(map[string]*Room),
		broadcast: make(chan Message),
	}
}

// HandleMessage memproses pesan dari broadcast channel dan mengirimkannya ke seluruh anggota room (kecuali pengirim).
func (s *Server) HandleMessage() {
	for {
		msg := <-s.broadcast

		if msg.sender.room == "" {
			continue
		}

		s.mutex.Lock()
		room := s.rooms[msg.sender.room]
		s.mutex.Unlock()

		for client := range room.Members {
			if client == msg.sender {
				continue
			}
			_, err := fmt.Fprintf(client.conn, "%s\n", msg.content)
			if err != nil {
				client.conn.Close()
				s.mutex.Lock()
				delete(s.clients, client)
				s.mutex.Unlock()
			}
		}
	}
}

// CreateRoom membuat room baru dengan client sebagai owner.
func (s *Server) CreateRoom(client *Client, roomName string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.rooms[roomName]; exists {
		fmt.Fprintf(client.conn, "Room %s sudah ada\n", roomName)
		return
	}

	s.rooms[roomName] = &Room{
		Name:    roomName,
		Owner:   client,
		Members: make(map[*Client]bool),
	}

	fmt.Fprintf(client.conn, "Room %s berhasil dibuat\n", roomName)
}

// DeleteRoom menghapus room jika client adalah owner-nya.
func (s *Server) DeleteRoom(client *Client, roomName string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	room, exists := s.rooms[roomName]
	if !exists {
		fmt.Fprintf(client.conn, "Room tidak ditemukan\n")
		return
	}

	if room.Owner != client {
		fmt.Fprintf(client.conn, "Anda bukan pemilik room %s\n", roomName)
		return
	}

	for member := range room.Members {
		member.room = ""
		fmt.Fprintf(member.conn, "[SERVER] Room %s telah dihapus\n", roomName)
	}

	delete(s.rooms, roomName)
	fmt.Fprintf(client.conn, "Room %s berhasil dihapus\n", roomName)
}

// ListRooms menampilkan daftar room yang tersedia kepada client.
func (s *Server) ListRooms(client *Client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.rooms) == 0 {
		fmt.Fprintf(client.conn,
			"Tidak ada room. Untuk membuat room, gunakan /create <nama_room>\n")
		return
	}

	fmt.Fprintf(client.conn, "Daftar room:\n")

	for _, room := range s.rooms {
		fmt.Fprintf(client.conn,
			"- %s | owner=%s | users=%d\n",
			room.Name,
			room.Owner.name,
			len(room.Members))
	}
}

// JoinRoom memindahkan client ke room yang dituju.
func (s *Server) JoinRoom(client *Client, roomName string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	room, exists := s.rooms[roomName]
	if !exists {
		fmt.Fprintf(client.conn, "Room tidak ditemukan\n")
		return
	}

	if client.room != "" {
		oldRoom := s.rooms[client.room]
		delete(oldRoom.Members, client)
	}

	room.Members[client] = true
	client.room = roomName

	//Memberi notifikasi kepada client yang ada di room tersebut
	for member := range room.Members {
		fmt.Fprintf(member.conn, "%s telah bergabung ke room %s\n", client.name, roomName)
	}
}

// LeaveRoom mengeluarkan client dari room saat ini.
func (s *Server) LeaveRoom(client *Client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if client.room == "" {
		fmt.Fprintf(client.conn, "Anda tidak berada di room manapun\n")
		return
	}

	room := s.rooms[client.room]
	delete(room.Members, client)

	oldRoom := client.room
	client.room = ""

	for c := range s.clients {
		fmt.Fprintf(c.conn, "%s keluar dari room %s\n", client.name, oldRoom)
	}
}

// WhoInRoom menampilkan daftar anggota di room client saat ini.
func (s *Server) WhoInRoom(client *Client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if client.room == "" {
		fmt.Fprintf(client.conn, "Anda belum masuk room\n")
		return
	}

	room := s.rooms[client.room]

	fmt.Fprintf(client.conn, "Anggota room %s:\n", room.Name)

	for member := range room.Members {
		label := ""
		if member == room.Owner {
			label = " (OWNER)"
		}
		fmt.Fprintf(client.conn, "- %s%s\n", member.name, label)
	}
}

// AddClient menambahkan client baru ke daftar client server.
func (s *Server) AddClient(client *Client) {
	s.mutex.Lock()
	s.clients[client] = true
	s.mutex.Unlock()
}

// RemoveClient menghapus client dari daftar client server.
func (s *Server) RemoveClient(client *Client) {
	s.mutex.Lock()
	delete(s.clients, client)
	s.mutex.Unlock()
}
