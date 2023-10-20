package main

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 4 * time.Second
	pingPeriod = 2 * time.Second
	pongWait   = 2 * pingPeriod
)

type Client struct {
	uuid string
	addr string
	conn *websocket.Conn
	send chan *Message
}

func (c *Client) Read() {
	defer func() {
		hub.leave <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(
		func(string) error {
			c.conn.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		},
	)

	for {
		m := &Message{}
		err := c.conn.ReadJSON(m)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("read error: %v", err)
			}
			break
		}

		m.client = c
		hub.message <- m
	}
}

func (c *Client) Write() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case m, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.conn.WriteJSON(m)
			if err != nil {
				log.Printf("write error: %v", err)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := c.conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				return
			}
		}
	}
}
