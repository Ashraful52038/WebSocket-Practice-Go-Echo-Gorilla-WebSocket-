package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // check origin as needed, for now allowing all origins
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// WebSocket Handler
func HandleWebSocket(hub *Hub) echo.HandlerFunc {
	return func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			log.Printf("Upgrade error: %v", err)
			return err
		}

		// Client ID generation (can be taken from Query parameter)
		clientID := c.QueryParam("client_id")
		if clientID == "" {
			clientID = generateClientID()
		}

		client := &Client{
			ID:   clientID,
			Conn: conn,
			Send: make(chan []byte, 256),
			Hub:  hub,
		}

		hub.register <- client

		// Read goroutine processing incoming messages from the client
		go func() {
			defer func() {
				hub.unregister <- client
				conn.Close()
			}()

			for {
				var msg WebSocketMessage
				err := conn.ReadJSON(&msg)
				if err != nil {
					log.Printf("Read error from %s: %v", client.ID, err)
					break
				}

				// route message based on its type
				switch msg.Type {
				case "unicast":
					hub.SendToClient(msg.To, []byte(msg.Content))

				case "broadcast":
					hub.broadcast <- []byte(msg.Content)

				case "multicast":
					hub.SendToRoom(msg.Room, []byte(msg.Content))

				case "join_room":
					hub.joinRoom <- JoinRoomRequest{
						Client:   client,
						RoomName: msg.Room,
					}

				case "leave_room":
					hub.leaveRoom <- LeaveRoomRequest{
						Client:   client,
						RoomName: msg.Room,
					}

				default:
					log.Printf("Unknown message type from %s: %s", client.ID, msg.Type)
					// inform the client about the unknown message type
					errorMsg := WebSocketMessage{
						Type:    "error",
						Content: "Unknown message type: " + msg.Type,
					}
					if err := conn.WriteJSON(errorMsg); err != nil {
						log.Printf("Error sending error message: %v", err)
					}
				}
			}
		}()

		// Write goroutine processing outgoing messages to the client
		go func() {
			for message := range client.Send {
				err := conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Printf("Write error to %s: %v", client.ID, err)
					break
				}
			}
		}()

		return nil
	}
}
