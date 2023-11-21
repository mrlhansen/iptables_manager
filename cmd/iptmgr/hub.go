package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Hub struct {
	uuid     string
	priority uint
	active   bool
	clients  map[string]*Client
	message  chan *Message
	join     chan *Client
	leave    chan *Client
	hosts    map[string]bool
	mu       sync.Mutex
}

var hub = &Hub{
	uuid:    uuid.NewString(),
	active:  true,
	message: make(chan *Message),
	join:    make(chan *Client),
	leave:   make(chan *Client),
	clients: make(map[string]*Client),
	hosts:   make(map[string]bool),
}

func (h *Hub) CheckPriority() {
	active := true
	for _, c := range h.clients {
		if c.priority > h.priority {
			active = false
			break
		}
	}
	if h.active != active {
		if active {
			log.Print("I am now active!")
		} else {
			log.Print("I am now backup!")
		}
		h.active = active
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.join:
			log.Printf("hub: joined: addr=%s uuid=%s priority=%d", c.addr, c.uuid, c.priority)
			h.clients[c.uuid] = c
			SendRegistryList(c)
			h.CheckPriority()
		case c := <-h.leave:
			close(c.send)
			delete(h.clients, c.uuid)
			if _, ok := h.hosts[c.addr]; ok {
				h.hosts[c.addr] = false
			}
			h.CheckPriority()
			log.Printf("hub: left: addr=%s uuid=%s priority=%d", c.addr, c.uuid, c.priority)
		case m := <-h.message:
			RecvMessage(m)
		}
	}
}

func (h *Hub) SetPriority(p uint) {
	h.priority = p
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
		"Instance-UUID":     []string{h.uuid},
		"Instance-Priority": []string{fmt.Sprint(h.priority)},
	}

	conn, resp, err := websocket.DefaultDialer.Dial(url, header)
	if resp != nil {
		if resp.StatusCode == StatusAlreadyConnected {
			delete(h.hosts, host)
		}
		if resp.StatusCode == StatusPriorityConflict {
			log.Fatalf("hub: priority conflict: host=%s", host)
		}
	}
	if err != nil {
		return
	}

	client := &Client{
		uuid:     resp.Header.Get("Instance-UUID"),
		priority: strToUint(resp.Header.Get("Instance-Priority")),
		addr:     host,
		conn:     conn,
		send:     make(chan *Message),
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
