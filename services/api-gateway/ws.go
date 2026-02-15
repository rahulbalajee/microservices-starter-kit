package main

import (
	"log"
	"net/http"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/util"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleRidersWebSocket(w http.ResponseWriter, r *http.Request) {
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

func handleDriversWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocker upgrade failed: %v\n", err)
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

	// TODO call drivers service

	msg := contracts.WSMessage{
		Type: "driver.cmd.register",
		Data: ,
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
		log.Printf("recieved message: %s\n", message)
	}
}
