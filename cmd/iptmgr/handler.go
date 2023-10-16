package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/mrlhansen/iptables_manager/pkg/iptables"
)

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

		id, err := iptables.CreateRuleSet(p.Rules)
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		}

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

		err = iptables.DeleteRuleSet(p.Id)
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		}

		return
	}

	http.Error(w, "", http.StatusBadRequest)
}

func ChainsHandler(w http.ResponseWriter, r *http.Request) {
	ok := Authenticate(w, r, AuthClient)
	if !ok {
		return
	}

	if r.Method == "POST" {
		return
	}

	if r.Method == "DELETE" {
		return
	}

	http.Error(w, "", http.StatusBadRequest)
}
