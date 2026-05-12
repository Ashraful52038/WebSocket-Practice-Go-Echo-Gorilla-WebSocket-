package main

import (
	"github.com/gorilla/websocket"
)

// Client স্ট্রাকচার
type Client struct {
	ID   string          // ইউনিক আইডি
	Conn *websocket.Conn // WebSocket সংযোগ
	Send chan []byte     // বার্তা পাঠানোর চ্যানেল
	Room string          // বর্তমান রুম (ঐচ্ছিক)
	Hub  *Hub            // হাবের রেফারেন্স
}
