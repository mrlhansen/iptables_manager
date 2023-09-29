package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/api/v1/iptables/rules", RulesHandler)
	err := http.ListenAndServe(":1234", nil)
	log.Fatal(err)
}
