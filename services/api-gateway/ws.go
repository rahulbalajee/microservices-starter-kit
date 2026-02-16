package main

import (
	"context"
	"log"
	"net/http"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/proto/driver"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
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
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

	if err := conn.WriteJSON(msg); err != nil {
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
