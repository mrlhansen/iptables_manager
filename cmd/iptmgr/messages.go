package main

import (
	"encoding/json"
	"log"

	"github.com/mrlhansen/iptables_manager/pkg/iptables"
	"github.com/mrlhansen/iptables_manager/pkg/registry"
)

const (
	MessageCreateRuleSet  = 1
	MessageDeleteRuleSet  = 2
	MessageRequestRuleSet = 3 // Response is CreateRuleSet
	MessageRegistryList   = 4
)

type Message struct {
	client *Client
	Type   int    `json:"type"`
	Epoch  int64  `json:"epoch,omitempty"`
	Name   string `json:"name,omitempty"`
	Data   []byte `json:"data,omitempty"`
}

func SendCreateRuleSet(id string, c *Client) {
	s := registry.Get(id)
	if s == nil {
		return
	}

	data, err := json.Marshal(s)
	if err != nil {
		return
	}

	m := &Message{
		Type:  MessageCreateRuleSet,
		Epoch: registry.GetEpoch(),
		Name:  id,
		Data:  data,
	}

	log.Printf("Messages: SendCreateRuleSet: id=%s", id)
	hub.SendOrBroadcast(m, c)
}

func SendDeleteRuleSet(id string, c *Client) {
	m := &Message{
		Type:  MessageDeleteRuleSet,
		Epoch: registry.GetEpoch(),
		Name:  id,
	}

	log.Printf("Messages: SendDeleteRuleSet: id=%s", id)
	hub.SendOrBroadcast(m, c)
}

func SendRequestRuleSet(id string, c *Client) {
	m := &Message{
		Type: MessageRequestRuleSet,
		Name: id,
	}

	log.Printf("Messages: SendRequestRuleSet: id=%s", id)
	hub.SendOrBroadcast(m, c)
}

func SendRegistryList(c *Client) {
	r := registry.List()
	data, err := json.Marshal(r)
	if err != nil {
		return
	}

	m := &Message{
		Type:  MessageRegistryList,
		Epoch: registry.GetEpoch(),
		Data:  data,
	}

	if c != nil {
		log.Printf("Messages: SendRegistryList: uuid=%s epoch=%d", c.uuid, m.Epoch)
	}
	hub.SendOrBroadcast(m, c)
}

func RecvCreateRuleSet(m *Message) {
	var s registry.Entry
	json.Unmarshal(m.Data, &s)

	log.Printf("Messages: RecvCreateRuleSet: id=%s", m.Name)
	err := iptables.CreateRuleSet(m.Name, s.Rule, m.Epoch)
	if err != nil {
		log.Printf("Failed: CreateRuleSet: %v", err)
		return
	}
}

func RecvDeleteRuleSet(m *Message) {
	log.Printf("Messages: RecvDeleteRuleSet: id=%s", m.Name)

	err := iptables.DeleteRuleSet(m.Name, m.Epoch)
	if err != nil {
		log.Printf("Failed: DeleteRuleSet: %v", err)
		return
	}
}

func RecvRequestRuleSet(m *Message) {
	log.Printf("Messages: RecvRequestRuleSet: id=%s", m.Name)
	SendCreateRuleSet(m.Name, m.client)
}

func RecvRegistryList(m *Message) {
	var remote, del, add []string

	log.Printf("Messages: RecvRegistryList: uuid=%s epoch=%d", m.client.uuid, m.Epoch)
	if m.Epoch <= registry.GetEpoch() {
		return
	}

	local := registry.List()
	json.Unmarshal(m.Data, &remote)

	// Items to delete from local
	for m := range local {
		found := false
		for n := range remote {
			if local[m] == remote[n] {
				found = true
				break
			}
		}
		if !found {
			del = append(del, local[m])
		}
	}

	// Items to fetch from remote
	for m := range remote {
		found := false
		for n := range local {
			if remote[m] == local[n] {
				found = true
				break
			}
		}
		if !found {
			add = append(add, remote[m])
		}
	}

	for n := range del {
		err := iptables.DeleteRuleSet(del[n], m.Epoch)
		if err != nil {
			log.Printf("Failed: DeleteRuleSet: %v", err)
		}
	}

	for n := range add {
		SendRequestRuleSet(add[n], m.client)
	}
}

func RecvMessage(m *Message) {
	switch m.Type {
	case MessageCreateRuleSet:
		RecvCreateRuleSet(m)
	case MessageDeleteRuleSet:
		RecvDeleteRuleSet(m)
	case MessageRequestRuleSet:
		RecvRequestRuleSet(m)
	case MessageRegistryList:
		RecvRegistryList(m)
	}
}
