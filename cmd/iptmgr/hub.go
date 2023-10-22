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
	message chan *Message
	join    chan *Client
	leave   chan *Client
	mu      sync.Mutex
	hosts   map[string]bool
}

var hub = &Hub{
	uuid:    uuid.NewString(),
	message: make(chan *Message),
	join:    make(chan *Client),
	leave:   make(chan *Client),
	clients: make(map[string]*Client),
	hosts:   make(map[string]bool),
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.join:
			log.Printf("hub: joined: addr=%s uuid=%s", c.addr, c.uuid)
			h.clients[c.uuid] = c
			SendRegistryList(c)
		case c := <-h.leave:
			close(c.send)
			delete(h.clients, c.uuid)
			if _, ok := h.hosts[c.addr]; ok {
				h.hosts[c.addr] = false
			}
			log.Printf("hub: left: addr=%s uuid=%s", c.addr, c.uuid)
		case m := <-h.message:
			RecvMessage(m)
		}
	}
}

func (h *Hub) Exists(uuid string) bool {
	_, ok := h.clients[uuid]
	return ok
}

func (h *Hub) SendOrBroadcast(m *Message, c *Client) {
	if c != nil {
		c.send <- m
		// select {
		// case c.send <- m:
		// default:
		// 	close(c.send)
		// 	delete(h.clients, c.uuid)
		// }
		return
	}
	for _, c := range h.clients {
		c.send <- m
		// select {
		// case c.send <- m:
		// default:
		// 	close(c.send)
		// 	delete(h.clients, c.uuid)
		// }
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
		send: make(chan *Message),
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
