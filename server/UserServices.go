package main

import (
	"net"
)

type User struct {
	conn     net.Conn
	username string
	room     *Room
}

func CheckUsername() {

}
