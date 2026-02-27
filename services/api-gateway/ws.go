package main

import (
	"context"
	"log"
	"net/http"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"
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

	c.SetPongHandler(func(string) error {
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

func (w *wsConn) Close() {
	w.CloseNormal()
	_ = w.c.Close()
}

func (w *wsConn) WriteJSON(v any) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.c.SetWriteDeadline(time.Now().Add(writeWait))
	return w.c.WriteJSON(v)
}

func (w *wsConn) ReadMessage() (int, []byte, error) {
	return w.c.ReadMessage()
}

func (w *wsConn) WriteMessage(messageType int, data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.c.SetWriteDeadline(time.Now().Add(writeWait))
	return w.c.WriteMessage(messageType, data)
}

var (
	connManager = messaging.NewConnectionManager()
)

func (app *application) handleRidersWebSocket(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		log.Println("no user id provided")
		return
	}

	conn, err := connManager.Upgrade(w, r)
	if err != nil {
		log.Printf("websocket upgrade failed: %v\n", err)
		return
	}

	ws := newWSConn(conn)
	defer ws.Close()

	connManager.Add(userID, ws)
	defer connManager.Remove(userID)

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			log.Printf("error reading message: %v\n", err)
			break
		}
		log.Printf("received message: %s\n", message)
	}
}

func (app *application) handleDriversWebSocket(w http.ResponseWriter, r *http.Request) {
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

	driverService := app.driverService.Load()
	if driverService == nil {
		http.Error(w, "driver service unavailable", http.StatusServiceUnavailable)
		return
	}

	conn, err := connManager.Upgrade(w, r)
	if err != nil {
		log.Printf("websocket upgrade failed: %v\n", err)
		return
	}

	ws := newWSConn(conn)
	defer ws.Close()

	connManager.Add(userID, ws)

	driverData, err := driverService.Client.RegisterDriver(
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
		connManager.Remove(userID)

		// r.Context() is already cancelled when the connection drops so use background instead to unregister drivers
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := driverService.Client.UnregisterDriver(
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

	if err := ws.WriteJSON(contracts.WSMessage{
		Type: contracts.DriverCmdRegister,
		Data: driverData.Driver,
	}); err != nil {
		log.Printf("error sending message: %v\n", err)
		return
	}

	// init queue consumer — cancelled when the WS read loop exits
	consumeCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	queues := []string{
		messaging.DriverCmdTripRequestQueue,
	}

	for _, queue := range queues {
		consumer := messaging.NewQueueConsumer(app.mq, connManager, queue)
		if err := consumer.Start(consumeCtx); err != nil {
			log.Printf("failed to start consumer for queue: %s: %v", queue, err)
		}
	}

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			log.Printf("error reading message: %v\n", err)
			break
		}
		log.Printf("received message: %s\n", message)
	}
}
