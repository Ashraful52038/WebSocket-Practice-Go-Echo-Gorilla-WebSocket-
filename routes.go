package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo, hub *Hub) {
	// WebSocket endpoint
	e.GET("/ws", HandleWebSocket(hub))

	// Static files (HTML)
	e.Static("/static", "static")
	e.File("/", "static/index.html")

	// API endpoints
	api := e.Group("/api")

	// Unicast via HTTP API
	api.POST("/send/:clientId", func(c echo.Context) error {
		clientID := c.Param("clientId")
		var req struct {
			Message string `json:"message"`
		}
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Invalid request",
			})
		}

		success := hub.SendToClient(clientID, []byte(req.Message))
		if !success {
			return c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Message: "Client not found",
			})
		}

		return c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "Message sent",
		})
	})

	// Multicast via HTTP API
	api.POST("/room/:roomName/send", func(c echo.Context) error {
		roomName := c.Param("roomName")
		var req struct {
			Message string `json:"message"`
		}
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Invalid request",
			})
		}

		count := hub.SendToRoom(roomName, []byte(req.Message))
		return c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Data: map[string]interface{}{
				"room":         roomName,
				"clients_sent": count,
			},
		})
	})

	// Get all connected clients
	api.GET("/clients", func(c echo.Context) error {
		clients := hub.GetConnectedClients()
		return c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Data: map[string]interface{}{
				"clients": clients,
				"count":   len(clients),
			},
		})
	})

	// Get room clients
	api.GET("/room/:roomName/clients", func(c echo.Context) error {
		roomName := c.Param("roomName")
		clients := hub.GetRoomClients(roomName)
		return c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Data: map[string]interface{}{
				"room":    roomName,
				"clients": clients,
				"count":   len(clients),
			},
		})
	})

	// Broadcast via HTTP API
	api.POST("/broadcast", func(c echo.Context) error {
		var req struct {
			Message string `json:"message"`
		}
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Invalid request",
			})
		}

		hub.broadcast <- []byte(req.Message)
		return c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "Broadcast sent",
		})
	})
}
