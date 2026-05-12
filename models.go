package main

import "time"

type WebSocketMessage struct {
	Type    string `json:"type"`    // "unicast", "broadcast", "multicast", "join_room", "leave_room"
	To      string `json:"to"`      // unicast -> clientID
	Room    string `json:"room"`    // multicast -> room name
	Content string `json:"content"` // message content
	Sender  string `json:"sender"`  // sender's ID
	Time    int64  `json:"time"`    // timestamp (optional)
}

// API Response Structure
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Helper Functions
func generateClientID() string {
	return "client_" + time.Now().Format("20060102150405") + "_" + randomString(6)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
