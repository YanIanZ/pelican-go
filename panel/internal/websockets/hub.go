package websockets

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

type Hub struct {
	mu      sync.RWMutex
	servers map[string]*ServerRoom
}

type ServerRoom struct {
	upstream    *websocket.Conn
	downstreams map[*websocket.Conn]bool
	mu          sync.Mutex
}

func NewHub() *Hub {
	return &Hub{servers: make(map[string]*ServerRoom)}
}

func (h *Hub) GetOrCreate(serverUUID string) *ServerRoom {
	h.mu.Lock()
	defer h.mu.Unlock()
	if room, ok := h.servers[serverUUID]; ok {
		return room
	}
	room := &ServerRoom{downstreams: make(map[*websocket.Conn]bool)}
	h.servers[serverUUID] = room
	return room
}

func (r *ServerRoom) AddDownstream(conn *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.downstreams[conn] = true
}

func (r *ServerRoom) RemoveDownstream(conn *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.downstreams, conn)
}

func (r *ServerRoom) Broadcast(msg []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for conn := range r.downstreams {
		conn.WriteMessage(websocket.TextMessage, msg)
	}
}
