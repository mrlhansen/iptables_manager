package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	var configFile string
	var dataDir string
	var listen string

	flag.StringVar(&configFile, "config", "config.yml", "Path to configuration file")
	flag.StringVar(&dataDir, "datadir", ".", "Path to persistent data storage")
	flag.StringVar(&listen, "listen", ":1234", "Listen address for web interface")
	flag.Parse()

	readConfig(configFile)
	log.Printf("registered configuration options")
	log.Printf("> config  = %s", configFile)
	log.Printf("> datadir = %s", dataDir)
	log.Printf("> listen  = %s", listen)

	// createChains()
	// e := iptables.PurgeChains()
	// log.Print(e, "heh")

	// registry.Init(".")

	http.HandleFunc("/api/v1/iptables/rules", RulesHandler)
	http.HandleFunc("/api/v1/iptables/chains", ChainsHandler)
	err := http.ListenAndServe(listen, nil)
	log.Fatal(err)
}
