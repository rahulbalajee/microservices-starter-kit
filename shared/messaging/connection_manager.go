package messaging

import (
	"errors"
	"log"
	"net/http"
	"ride-sharing/shared/contracts"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	ErrConnectionNotFound = errors.New("connection not found")
)

type WSConn interface {
	WriteJSON(v any) error
}

type ConnectionManager struct {
	connections map[string]WSConn
	upgrader    websocket.Upgrader
	mu          sync.RWMutex
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]WSConn),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (cm *ConnectionManager) Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	return cm.upgrader.Upgrade(w, r, nil)
}

func (cm *ConnectionManager) Add(id string, conn WSConn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.connections[id] = conn
	log.Printf("added connection for user: %s", id)
}

func (cm *ConnectionManager) Remove(id string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	_, exists := cm.connections[id]
	if !exists {
		return ErrConnectionNotFound
	}
	delete(cm.connections, id)
	return nil
}

func (cm *ConnectionManager) Get(id string) (WSConn, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	conn, exists := cm.connections[id]
	if !exists {
		return nil, ErrConnectionNotFound
	}
	return conn, nil
}

func (cm *ConnectionManager) SendMessage(id string, message contracts.WSMessage) error {
	cm.mu.RLock()
	conn, exists := cm.connections[id]
	cm.mu.RUnlock()

	if !exists {
		return ErrConnectionNotFound
	}

	return conn.WriteJSON(message)
}
