package main

import (
	"log"
	"net/http"

	"github.com/mrlhansen/iptables_manager/pkg/registry"
)

func main() {
	registry.Init(".")

	http.HandleFunc("/api/v1/iptables/rules", RulesHandler)
	err := http.ListenAndServe(":1234", nil)
	log.Fatal(err)
}
