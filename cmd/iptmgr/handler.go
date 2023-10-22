package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/mrlhansen/iptables_manager/pkg/iptables"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const (
	AuthNone   int = 1
	AuthClient int = 2
	AuthServer int = 3
)

type RulesPostRequest struct {
	Rules []iptables.Rule `json:"rules"`
}

type RulesPostResponse struct {
	Id string `json:"id"`
}

type RulesDeleteRequest struct {
	Id string `json:"id"`
}

func Authenticate(w http.ResponseWriter, r *http.Request, a int) bool {
	log.Printf("%s request from %s to %s", r.Method, r.RemoteAddr, r.URL)
	w.Header().Set("Content-Type", "application/json")

	if a == AuthNone {
		return true
	}

	if a == AuthServer {
		uuid := r.Header.Get("Instance-UUID")
		if len(uuid) != 36 {
			http.Error(w, "", http.StatusBadRequest)
			return false
		}
	}

	// auth := r.Header.Get("Authorization")
	// token := strings.TrimPrefix(auth, "Bearer ")
	// if token != "" {
	// 	http.Error(w, "", http.StatusUnauthorized)
	// 	return false
	// }

	return true
}

func RulesHandler(w http.ResponseWriter, r *http.Request) {
	ok := Authenticate(w, r, AuthClient)
	if !ok {
		return
	}

	if r.Method == "POST" {
		p := RulesPostRequest{}
		err := json.NewDecoder(r.Body).Decode(&p)
		if (err != nil) || (len(p.Rules) == 0) {
			http.Error(w, fmt.Sprint(err), http.StatusBadRequest)
			return
		}

		// Create rules
		id, rules, err := iptables.PrepareRuleSet(p.Rules)
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		}

		err = iptables.CreateRuleSet(id, rules, 0)
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		}

		// Broadcast
		go SendCreateRuleSet(id, nil)

		// Response
		q := RulesPostResponse{
			Id: id,
		}
		json.NewEncoder(w).Encode(q)

		return
	}

	if r.Method == "DELETE" {
		p := RulesDeleteRequest{}
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusBadRequest)
			return
		}

		// Delete rules
		err = iptables.DeleteRuleSet(p.Id, 0)
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		}

		// Broadcast
		go SendDeleteRuleSet(p.Id, nil)

		return
	}

	http.Error(w, "", http.StatusBadRequest)
}

func ClusterHandler(w http.ResponseWriter, r *http.Request) {
	hub.Lock()
	defer hub.Unlock()

	ok := Authenticate(w, r, AuthServer)
	if !ok {
		return
	}

	uuid := r.Header.Get("Instance-UUID")
	if hub.Exists(uuid) {
		http.Error(w, "", http.StatusConflict)
		return
	}

	header := http.Header{
		"Instance-UUID": []string{hub.uuid},
	}
	conn, err := upgrader.Upgrade(w, r, header)
	if err != nil {
		log.Printf("handler: upgrade failed: %v", err)
		return
	}

	client := &Client{
		uuid: uuid,
		addr: r.RemoteAddr,
		conn: conn,
		send: make(chan *Message),
	}

	hub.join <- client
	go client.Read()
	go client.Write()
}
