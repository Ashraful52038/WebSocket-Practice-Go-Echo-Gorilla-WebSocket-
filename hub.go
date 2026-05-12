package main

import (
	"log"
	"sync"
)

// Room structure
type Room struct {
	Name    string
	Clients map[string]*Client
	mu      sync.RWMutex
}

// Hub structure
type Hub struct {
	// Unicast & Broadcast
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte

	// Multicast (Room)
	rooms      map[string]*Room
	createRoom chan string
	joinRoom   chan JoinRoomRequest
	leaveRoom  chan LeaveRoomRequest

	mu sync.RWMutex
}

type JoinRoomRequest struct {
	Client   *Client
	RoomName string
}

type LeaveRoomRequest struct {
	Client   *Client
	RoomName string
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),

		rooms:      make(map[string]*Room),
		createRoom: make(chan string),
		joinRoom:   make(chan JoinRoomRequest),
		leaveRoom:  make(chan LeaveRoomRequest),
	}
}

// for handling all hub operations (register, unregister, broadcast, room management)
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			log.Printf("✅ Client %s registered. Total: %d", client.ID, len(h.clients))
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
				log.Printf("❌ Client %s unregistered. Total: %d", client.ID, len(h.clients))
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			log.Printf("📢 Broadcasting to %d clients", len(h.clients))
			for _, client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client.ID)
				}
			}
			h.mu.RUnlock()

		case roomName := <-h.createRoom:
			h.createRoomHandler(roomName)

		case req := <-h.joinRoom:
			h.joinRoomHandler(req.Client, req.RoomName)

		case req := <-h.leaveRoom:
			h.leaveRoomHandler(req.Client, req.RoomName)
		}
	}
}

// helper methods for room management and message routing
func (h *Hub) createRoomHandler(roomName string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.rooms[roomName]; !exists {
		h.rooms[roomName] = &Room{
			Name:    roomName,
			Clients: make(map[string]*Client),
		}
		log.Printf("🏠 Room created: %s", roomName)
	}
}

func (h *Hub) joinRoomHandler(client *Client, roomName string) {
	h.mu.Lock()
	room, exists := h.rooms[roomName]
	if !exists {
		room = &Room{
			Name:    roomName,
			Clients: make(map[string]*Client),
		}
		h.rooms[roomName] = room
		log.Printf("🏠 Auto-created room: %s", roomName)
	}
	h.mu.Unlock()

	room.mu.Lock()
	room.Clients[client.ID] = client
	client.Room = roomName
	room.mu.Unlock()

	log.Printf("🚪 Client %s joined room: %s (Total in room: %d)",
		client.ID, roomName, len(room.Clients))
}

func (h *Hub) leaveRoomHandler(client *Client, roomName string) {
	h.mu.RLock()
	room, exists := h.rooms[roomName]
	h.mu.RUnlock()

	if exists {
		room.mu.Lock()
		delete(room.Clients, client.ID)
		room.mu.Unlock()

		if client.Room == roomName {
			client.Room = ""
		}
		log.Printf("🚪 Client %s left room: %s", client.ID, roomName)
	}
}

// Unicast: send message to a specific client by ID
func (h *Hub) SendToClient(clientID string, message []byte) bool {
	h.mu.RLock()
	client, exists := h.clients[clientID]
	h.mu.RUnlock()

	if !exists {
		log.Printf("⚠️ Client %s not found", clientID)
		return false
	}

	select {
	case client.Send <- message:
		log.Printf("🔵 Unicast to %s: %s", clientID, string(message))
		return true
	default:
		log.Printf("❌ Failed to send to %s (buffer full)", clientID)
		return false
	}
}

// Multicast: send message to all clients in a specific room
func (h *Hub) SendToRoom(roomName string, message []byte) int {
	h.mu.RLock()
	room, exists := h.rooms[roomName]
	h.mu.RUnlock()

	if !exists {
		log.Printf("⚠️ Room %s not found", roomName)
		return 0
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	sentCount := 0
	log.Printf("🟢 Multicast to room '%s' (%d clients): %s",
		roomName, len(room.Clients), string(message))

	for _, client := range room.Clients {
		select {
		case client.Send <- message:
			sentCount++
		default:
			log.Printf("⚠️ Failed to send to %s in room %s", client.ID, roomName)
		}
	}

	return sentCount
}

// GetConnectedClients returns a list of all connected client IDs
func (h *Hub) GetConnectedClients() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := make([]string, 0, len(h.clients))
	for id := range h.clients {
		clients = append(clients, id)
	}
	return clients
}

// GetRoomClients returns a list of all clients in a specific room
func (h *Hub) GetRoomClients(roomName string) []string {
	h.mu.RLock()
	room, exists := h.rooms[roomName]
	h.mu.RUnlock()

	if !exists {
		return []string{}
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	clients := make([]string, 0, len(room.Clients))
	for id := range room.Clients {
		clients = append(clients, id)
	}
	return clients
}
