package main

type Room struct {
	roomName string
	users    map[*User]bool
}

var rooms []Room

func NewRoom(name string) *Room {
	if rooms == nil {
		rooms = make([]Room, 1)
	}

	newRoom := Room{
		roomName: name,
		users:    make(map[*User]bool),
	}

	rooms = append(rooms, newRoom)

	return &newRoom
}

func (cg *Room) Broadcast(message string, sender User) {

}
