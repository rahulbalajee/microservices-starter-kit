package main

import (
	"context"
	"log"
	"net/http"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/proto/driver"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	pongWait       = time.Minute
	pingPeriod     = (pongWait * 9) / 10
	writeWait      = 10 * time.Second
	maxMessageSize = 1 << 20
)

type wsConn struct {
	c    *websocket.Conn
	mu   sync.Mutex
	stop chan struct{}
	done chan struct{}
}

func newWSConn(c *websocket.Conn) *wsConn {
	c.SetReadLimit(maxMessageSize)
	c.SetReadDeadline(time.Now().Add(pongWait))

	c.SetPongHandler(func(appData string) error {
		c.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	w := &wsConn{
		c:    c,
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}

	go w.pingLoop()

	return w
}

func (w *wsConn) pingLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	defer close(w.done)

	for {
		select {
		case <-ticker.C:
			// single writer + per-write deadline
			_ = w.WriteMessage(websocket.PingMessage, nil)
		case <-w.stop:
			return
		}
	}
}

func (w *wsConn) WriteJSON(v any) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.c.SetWriteDeadline(time.Now().Add(writeWait))
	return w.c.WriteJSON(v)
}

func (w *wsConn) WriteMessage(messageType int, data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.c.SetWriteDeadline(time.Now().Add(writeWait))
	return w.c.WriteMessage(messageType, data)
}

func (w *wsConn) CloseNormal() {
	// stop ping loop
	select {
	case <-w.stop:
		// already closed
	default:
		close(w.stop)
	}
	<-w.done

	// best-effort close handshake
	_ = w.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	)
}

var upgrader = websocket.Upgrader{
	// TODO: Fix this before production
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (app *application) handleRidersWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v\n", err)
		return
	}
	defer conn.Close()

	ws := newWSConn(conn)
	defer ws.CloseNormal()

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		log.Println("no user id provided")
		return
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("error reading message: %v\n", err)
			break
		}
		log.Printf("received message: %s\n", message)
	}
}

func (app *application) handleDriversWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v\n", err)
		return
	}
	defer conn.Close()

	ws := newWSConn(conn)
	defer ws.CloseNormal()

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		log.Println("no user id provided")
		return
	}

	packageSlug := r.URL.Query().Get("packageSlug")
	if packageSlug == "" {
		log.Println("no package slug provided")
		return
	}

	driverData, err := app.driverService.Load().Client.RegisterDriver(
		r.Context(),
		&driver.RegisterDriverRequest{
			DriverID:    userID,
			PackageSlug: packageSlug,
		},
	)
	if err != nil {
		log.Printf("error registering driver %v\n", err)
		return
	}

	defer func() {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		_, err := app.driverService.Load().Client.UnregisterDriver(
			ctx,
			&driver.RegisterDriverRequest{
				DriverID:    userID,
				PackageSlug: packageSlug,
			},
		)
		if err != nil {
			log.Printf("error unregistering driver %v\n", err)
			return
		}
		log.Println("driver unregistered: ", userID)
	}()

	msg := contracts.WSMessage{
		Type: "driver.cmd.register",
		Data: driverData.Driver,
	}

	if err := ws.WriteJSON(msg); err != nil {
		log.Printf("error sending message: %v\n", err)
		return
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("error reading message: %v\n", err)
			break
		}
		log.Printf("received message: %s\n", message)
	}
}
