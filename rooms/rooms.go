package rooms

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Room struct {
	Peers map[*websocket.Conn]bool
	mu sync.Mutex
}

type SignalMessage struct {
	Type string `json:"type"`
	To string `json:"to,omitempty"`
	From string `json:"from,omitempty"`
	Payload json.RawMessage `json:"payload"`
}

var(
	rooms = make(map[string]*Room)
	mu sync.Mutex
)

func CreateRoom() (string,error) {
	roomID := uuid.NewString()
	mu.Lock()
	_, exists  := rooms[roomID]
	if !exists {
		room := &Room{
			Peers: make(map[*websocket.Conn]bool),
		}

		rooms[roomID] = room
	}
	mu.Unlock()

	return roomID, nil
}

func GetRoom(roomID string) (*Room, bool) {
	room, exists  := rooms[roomID]
	return room, exists
}

func GetRooms() ([]map[string]interface{}, error) {
	mu.Lock()
	roomList := make([]map[string]interface{}, 0)
	for id, room := range rooms {
		room.mu.Lock()
		roomList = append(roomList, map[string]interface{}{
			"id": id,
			"peer_count": len(room.Peers),
		})

		room.mu.Unlock()
	}

	mu.Unlock()
	return roomList, nil
}

func BroadcastToPeers(room *Room, sender *websocket.Conn, msg SignalMessage) {
	room.mu.Lock()
	defer room.mu.Unlock()

	for peer := range room.Peers {
		if peer != sender {
			err := peer.WriteJSON(msg)
			if err != nil {
				log.Printf("Broadcast error: %v", err)
			}
		}
	}
}

func Connect(room *Room, conn *websocket.Conn) error{
	room.mu.Lock()
	room.Peers[conn] = true
	room.mu.Unlock()

	return nil
}

func RemovePeer(room *Room, conn *websocket.Conn){
	room.mu.Lock()
	delete(room.Peers, conn)
	room.mu.Unlock()
}

func DeleteRoom(roomID string) {
	mu.Lock()
	delete(rooms, roomID)
	mu.Unlock()
}