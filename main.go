package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var (
	rooms = make(map[string]*Room)
	mu sync.Mutex
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
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


func main(){
	e := echo.New()
	// e.Use(middleware.Logger())
	// e.Use(middleware.Recover())
	// e.Use(middleware.CORS())

    e.GET("/ws", handleWebSocket)
    e.GET("/rooms", listRooms)
    e.GET("/health", healthCheck)

    // Start server
    log.Println("Starting signaling server on :8080")
    e.Logger.Fatal(e.Start(":8080"))

}

func listRooms(c echo.Context) error {
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
	return c.JSON(http.StatusOK, roomList)
}

func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string {
		"status": "healthy",
	})
}

func handleWebSocket(c echo.Context) error{
	roomID := c.QueryParam("room")
	if roomID == ""{
		return c.String(http.StatusBadRequest, "Room ID required")
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return err
	}

	defer conn.Close()

	mu.Lock()
	room, exists  := rooms[roomID]
	if !exists {
		room = &Room{
			Peers: make(map[*websocket.Conn]bool),
		}

		rooms[roomID] = room
	}
	mu.Unlock()

	room.mu.Lock()
	room.Peers[conn] = true
	peerCount := len(room.Peers)
	room.mu.Unlock()

	log.Printf("New peer joined the room %s. Total peers: %d", roomID, peerCount)

	broadcastToPeers(room, conn, SignalMessage{
		Type: "peer_joined",
		From: conn.RemoteAddr().String(),
	})

	// handle WebSocket messages
	for {
		var msg SignalMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}

		log.Printf("Room %s: Received message type: %s from %s", roomID, msg.Type, conn.RemoteAddr().String())
		switch msg.Type {
		case "offer", "answer", "ice-candidate":
			room.mu.Lock()
			for peer := range room.Peers {
				if peer != conn {
					msg.From = conn.RemoteAddr().String()
					err := peer.WriteJSON(msg)
					if err != nil {
						log.Printf("Write  error: %v", err)
					}
				}
			}
			room.mu.Unlock()
		}
		
	}

	room.mu.Lock()
	delete(room.Peers, conn)
	remainingPeers := len(room.Peers)
	room.mu.Unlock()

	log.Printf("Peer left room %s. Remaining peers: %d", roomID, remainingPeers)

	// Clean up empty rooms
	if remainingPeers == 0 {
		mu.Lock()
		delete(rooms, roomID)
		mu.Unlock()
		log.Printf("Room %s deleted as it is empty", roomID)
	}

	broadcastToPeers(room, conn, SignalMessage{
		Type: "peer_left",
		From: conn.RemoteAddr().String(),
	})

	return nil
}

func broadcastToPeers(room *Room, sender *websocket.Conn, msg SignalMessage) {
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