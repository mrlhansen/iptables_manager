package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Hub struct {
	uuid    string
	clients map[string]*Client
	message chan []byte // should this channel be a struct of client + byte array? Then we can easily respond to a given client if needed. first part of message can also be the uuid (36 chars)
	join    chan *Client
	leave   chan *Client
	mu      sync.Mutex
	hosts   map[string]bool
}

var hub = &Hub{
	uuid:    uuid.NewString(),
	message: make(chan []byte),
	join:    make(chan *Client),
	leave:   make(chan *Client),
	clients: make(map[string]*Client),
	hosts:   make(map[string]bool),
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.join:
			h.clients[client.uuid] = client
			log.Printf("hub: client joined: addr=%s uuid=%s", client.addr, client.uuid)
		case client := <-h.leave:
			close(client.send)
			delete(h.clients, client.uuid)
			if _, ok := h.hosts[client.addr]; ok {
				h.hosts[client.addr] = false
			}
			log.Printf("hub: client left: addr=%s uuid=%s", client.addr, client.uuid)
		case message := <-h.message:
			log.Printf("hub: message received: %s", message)
			RecvMessage(message)
		}
	}
}

func (h *Hub) Exists(uuid string) bool {
	_, ok := h.clients[uuid]
	return ok
}

func (h *Hub) Broadcast(message []byte) {
	for _, client := range h.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client.uuid)
		}
	}
}

func (h *Hub) Connect(host string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	url := "ws://" + host + "/api/v1/cluster"
	header := http.Header{
		"Instance-UUID": []string{h.uuid},
	}

	conn, resp, err := websocket.DefaultDialer.Dial(url, header)
	if resp != nil {
		if resp.StatusCode == http.StatusConflict {
			delete(h.hosts, host)
		}
	}
	if err != nil {
		return
	}

	client := &Client{
		uuid: resp.Header.Get("Instance-UUID"),
		addr: host,
		conn: conn,
		send: make(chan []byte),
	}

	h.hosts[host] = true
	hub.join <- client
	go client.Read()
	go client.Write()
}

func (h *Hub) Reconnect() {
	for {
		c := h.hosts
		if len(c) == 0 {
			return
		}
		for host, connected := range c {
			if !connected {
				hub.Connect(host)
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func (h *Hub) Lock() {
	h.mu.Lock()
}

func (h *Hub) Unlock() {
	h.mu.Unlock()
}
