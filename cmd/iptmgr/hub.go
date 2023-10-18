package main

import (
	"log"
)

type Hub struct {
	clients map[*Client]bool
	message chan []byte
	join    chan *Client
	leave   chan *Client
}

var hub = &Hub{
	message: make(chan []byte),
	join:    make(chan *Client),
	leave:   make(chan *Client),
	clients: make(map[*Client]bool),
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.join:
			h.clients[client] = true
			log.Printf("hub: client joined: %s", client.addr)
		case client := <-h.leave:
			close(client.send)
			delete(h.clients, client)
			log.Printf("hub: client left: %s", client.addr)
		case message := <-h.message:
			log.Printf("hub: message received: %s", message)
		}
	}
}

func (h *Hub) broadcast(message []byte) {
	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}
