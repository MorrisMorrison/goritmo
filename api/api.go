package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MorrisMorrison/goritmo/rooms"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func ListRooms(c echo.Context) error {
	roomList,err := rooms.GetRooms()
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	return c.JSON(http.StatusOK, roomList)
}

func HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string {
		"status": "healthy",
	})
}

func CreateRoom(c echo.Context) error{
	roomID, err := rooms.CreateEmptyRoom()
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, map[string]string {
		"roomID": roomID,
	})	
}

func HandleWebSocket(c echo.Context) error{
	roomID := c.QueryParam("room")

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return err
	}

	defer conn.Close()

	var room *rooms.Room
	var exists bool
	if roomID != "" {
		room, exists = rooms.GetRoom(roomID)
		if !exists {
			return fmt.Errorf("Could not find room.")
		}
	}else{
		roomID, err = rooms.CreateRoom(conn)
		if (err != nil) {
			return err
		}

		room, exists = rooms.GetRoom(roomID)
		if !exists {
			return fmt.Errorf("Could not find room.")
		}
	}

	rooms.Connect(room, conn)
	if (err != nil) {
		return err
	}

	totalPeers := len(room.Peers)
	log.Printf("New peer joined the room %s. Total peers: %d", roomID, totalPeers)

	rooms.MessagePeers(room, conn, rooms.SignalMessage{
		Type: "peer_joined",
		From: conn.RemoteAddr().String(),
	})

	for {
		var msg rooms.SignalMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}

		log.Printf("Room %s: Received message type: %s from %s", roomID, msg.Type, conn.RemoteAddr().String())
		switch msg.Type {
		case "offer", "answer", "ice-candidate":
			rooms.MessagePeers(room, conn, msg)
		case "viewer-join":
			rooms.MessageOwner(room, conn, rooms.SignalMessage{
				Type: "offer_request",
				To: conn.RemoteAddr().String(),
			})
		}
	}

	rooms.RemovePeer(room, conn)
	remainingPeers := len(room.Peers)
	log.Printf("Peer left room %s. Remaining peers: %d", roomID, remainingPeers)

	if remainingPeers == 0 {
		rooms.DeleteRoom(roomID)
	}

	rooms.MessagePeers(room, conn, rooms.SignalMessage{
		Type: "peer_left",
		From: conn.RemoteAddr().String(),
	})

	return nil
}

