# WebSocket Practice (Go + Echo + Gorilla WebSocket)

A small real-time messaging demo built with **Go**, **Echo** (HTTP server) and **Gorilla WebSocket**.

It supports:

- **WebSocket messaging** with JSON payloads
  - `unicast` (private message to a specific client)
  - `broadcast` (message to all connected clients)
  - `multicast` (message to a room)
  - `join_room` / `leave_room`
- **HTTP API** to drive messaging and inspect connected clients
  - Send unicast/broadcast/multicast via REST
  - List online clients and room members
- A simple browser UI at `/`.

---

## Project layout

- `main.go` – server bootstrap
- `routes.go` – HTTP routes (static files + REST + websocket endpoint)
- `handler.go` – `/ws` websocket handler
- `hub.go` – in-memory hub, client registry, room management
- `client.go` – client structure
- `models.go` – message/response types and client id generator
- `static/index.html` – browser UI

---

## Requirements

- Go (the project uses Go **1.26.2** in `go.mod`)
- No database (in-memory only)

---

## Run

```bash
go run .
```

Server will start on:

- HTTP: `http://localhost:8080/`
- WebSocket: `ws://localhost:8080/ws`

Open **http://localhost:8080** in multiple browser tabs to see:

- broadcasting to everyone
- unicast to a specific client id
- room based messaging

---

## Workflow (high-level)

```mermaid
graph TD
  UI[Browser UI
 static/index.html] -->|1) send JSON| WS[WebSocket /ws]
  WS -->|2) routes by type| HUB[Hub (hub.go)
 clients map + rooms]

  HUB -->|3a) unicast| C1[Target Client Send channel]
  HUB -->|3b) broadcast| ALL[All Client Send channels]
  HUB -->|3c) multicast| ROOM[Room members]

  C1 -->|4) websocket frame| UI2[All browsers/clients]
  ALL --> UI2
  ROOM --> UI2

  UI -->|HTTP poll| API[Echo REST API (/api/clients)]
  API --> HUB
```

## WebSocket protocol


Connect to:

- `GET /ws?client_id=<id>`

If `client_id` is not provided, the server generates one.

### Client -> Server message (JSON)

```json
{
  "type": "unicast|broadcast|multicast|join_room|leave_room",
  "to": "clientId",        
  "room": "roomName",
  "content": "message text",
  "sender": "senderId",
  "time": 1710000000000
}
```

Only the relevant fields are needed for each `type`:

- **unicast**: use `to` + `content`
- **broadcast**: use `content`
- **multicast**: use `room` + `content`
- **join_room**: use `room`
- **leave_room**: use `room`

### Server -> Client message

The server sends **raw text** (string) as websocket frames (wrapped in `client.Send`).

Unknown message types result in an error JSON:

```json
{ "type": "error", "content": "Unknown message type: ..." }
```

> Note: The provided frontend UI formats/infers message types based on message content (e.g., it looks for markers like `[Private` or `[Room`).

---

## HTTP API

All API endpoints are under `/api`.

All responses follow:

```json
{ "success": true|false, "message": "...", "data": { } }
```

### Send unicast

`POST /api/send/:clientId`

Body:

```json
{ "message": "hello" }
```

### Broadcast

`POST /api/broadcast`

Body:

```json
{ "message": "announcement" }
```

### Multicast to room

`POST /api/room/:roomName/send`

Body:

```json
{ "message": "room message" }
```

Returns how many clients were sent.

### List connected clients

`GET /api/clients`

### List room clients

`GET /api/room/:roomName/clients`

---

## Frontend UI

The file `static/index.html` provides an in-browser demo that:

- connects automatically to `ws://localhost:8080/ws?client_id=...`
- auto-joins a default room named `general`
- lets you:
  - send unicast (private)
  - send broadcast
  - join/leave rooms
  - send room messages
- polls online clients via `GET /api/clients`

---

## Notes / limitations

- All data is **in-memory**. Restarting the server clears clients/rooms.
- `CheckOrigin` is currently permissive (`true`) in the websocket upgrader—tighten this for production.
- Message routing is done via the hub using buffered channels (`Send` channel size is 256).

---

## License

Add your license here (if applicable).

# WebSocket-Practice-Go-Echo-Gorilla-WebSocket-
