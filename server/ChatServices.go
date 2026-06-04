package main

import (
	"net"
)

type Room struct {
	users map[User]bool
}

type User struct {
	conn      net.Conn
	username  string
	chatGroup *Room
}

func NewRoom() *Room {
	return &Room{
		users: make(map[User]bool),
	}
}

func (cg *Room) Broadcast(message string, sender User) {

}
