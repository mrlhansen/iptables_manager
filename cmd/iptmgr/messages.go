package main

import (
	"encoding/json"
	"log"

	"github.com/mrlhansen/iptables_manager/pkg/iptables"
	"github.com/mrlhansen/iptables_manager/pkg/registry"
)

const (
	MessageCreateRuleSet       = 1
	MessageDeleteRuleSet       = 2
	MessageRequestRuleSet      = 3 // Response is CreateRuleSet
	MessageSynchronizeRegistry = 4 // Rename to MessageRegistryList ?
)

type Message struct {
	Type int    `json:"type"`
	Name string `json:"name,omitempty"`
	Data []byte `json:"data,omitempty"`
}

func SendCreateRuleSet(id string) {
	s := registry.Get(id)
	if s == nil {
		return
	}

	data, err := json.Marshal(s)
	if err != nil {
		return
	}

	m := Message{
		Type: MessageCreateRuleSet,
		Name: id,
		Data: data,
	}

	msg, err := json.Marshal(m)
	if err != nil {
		return
	}

	log.Printf("Broadcasting CreateRuleSet: %s", id)
	hub.Broadcast(msg)
}

func SendDeleteRuleSet(id string) {
	m := Message{
		Type: MessageDeleteRuleSet,
		Name: id,
	}

	msg, err := json.Marshal(m)
	if err != nil {
		return
	}

	log.Printf("Broadcasting DeleteRuleSet: %s", id)
	hub.Broadcast(msg)
}

func RecvCreateRuleSet(m *Message) {
	var s registry.Entry
	json.Unmarshal(m.Data, &s)

	log.Printf("Receiving CreateRuleSet: %s", m.Name)
	err := iptables.CreateRuleSet(m.Name, s.Rule)
	if err != nil {
		log.Printf("Failed: CreateRuleSet: %v", err)
		return
	}
}

func RecvDeleteRuleSet(m *Message) {
	log.Printf("Receiving DeleteRuleSet: %s", m.Name)

	err := iptables.DeleteRuleSet(m.Name)
	if err != nil {
		log.Printf("Failed: DeleteRuleSet: %v", err)
		return
	}
}

func SendSynchronizeRegistry(uuid string) { // argument could be a client

}

func RecvSynchronizeRegistry(m *Message) {

}

func RecvMessage(data []byte) {
	var m Message
	json.Unmarshal(data, &m)

	switch m.Type {
	case MessageCreateRuleSet:
		RecvCreateRuleSet(&m)
	case MessageDeleteRuleSet:
		RecvDeleteRuleSet(&m)
	case MessageSynchronizeRegistry:
		RecvSynchronizeRegistry(&m)
	}
}
