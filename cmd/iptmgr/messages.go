package main

import (
	"encoding/json"
	"log"

	"github.com/mrlhansen/iptables_manager/pkg/iptables"
	"github.com/mrlhansen/iptables_manager/pkg/registry"
)

const (
	RuleAppend = 1
	RuleDelete = 2
	SyncRules  = 3
)

type Message struct {
	Type int    `json:"type"`
	Name string `json:"name,omitempty"`
	Data []byte `json:"data,omitempty"`
}

func SendRuleAppend(id string) {
	s := registry.Get(id)
	if s == nil {
		return
	}

	data, err := json.Marshal(s)
	if err != nil {
		return
	}

	m := Message{
		Type: RuleAppend,
		Name: id,
		Data: data,
	}

	msg, err := json.Marshal(m)
	if err != nil {
		return
	}

	log.Printf("broadcasting rule append: %s", id)
	hub.broadcast(msg)
}

func SendRuleDelete(id string) {
	m := Message{
		Type: RuleDelete,
		Name: id,
	}

	msg, err := json.Marshal(m)
	if err != nil {
		return
	}

	log.Printf("broadcasting rule delete: %s", id)
	hub.broadcast(msg)
}

func RecvRuleAppend(m *Message) {
	var s registry.Entry
	json.Unmarshal(m.Data, &s)

	log.Printf("receiving rule append: %s", m.Name)

	err := iptables.CreateRuleSet(m.Name, s.Rule)
	if err != nil {
		log.Print("failed to create rule")
		return
	}
}

func RecvRuleDelete(m *Message) {
	log.Printf("receiving rule delete: %s", m.Name)
	err := iptables.DeleteRuleSet(m.Name)
	if err != nil {
		log.Print("failed to delete rule")
		return
	}
}

func RecvMessage(data []byte) {
	var m Message
	json.Unmarshal(data, &m)

	switch m.Type {
	case RuleAppend:
		RecvRuleAppend(&m)
	case RuleDelete:
		RecvRuleDelete(&m)
	}
}
