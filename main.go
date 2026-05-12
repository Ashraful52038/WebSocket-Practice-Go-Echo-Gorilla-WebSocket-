package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// hub created and run in a separate goroutine
	hub := NewHub()
	go hub.Run()

	SetupRoutes(e, hub)

	log.Println("🚀 Server starting on :8080")
	log.Println("📡 WebSocket endpoint: ws://localhost:8080/ws")
	log.Println("🌐 Open http://localhost:8080 in your browser")

	if err := e.Start(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
