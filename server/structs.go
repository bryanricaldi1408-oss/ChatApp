package main

import "net"

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

type Message struct {
	sender  *Client
	content string
}
