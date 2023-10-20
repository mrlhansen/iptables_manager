package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/mrlhansen/iptables_manager/pkg/iptables"
	"github.com/mrlhansen/iptables_manager/pkg/registry"
)

func flagNotPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return !found
}

func onExit(purge bool) {
	// we might want to expand this function to also remove static rules and all rules (not just chains)
	// gracefully close connections to clients
	if purge {
		err := iptables.PurgeChains()
		if err != nil {
			log.Print(err)
		}
	}
}

func main() {
	var configFile string
	var dataPath string
	var logFile string
	var listen string
	var peers string
	var purgeOnExit bool

	// Check for root or iptables permissions?

	flag.StringVar(&configFile, "config-file", "config.yml", "Path to configuration file")
	flag.StringVar(&dataPath, "data-path", ".", "Path to persistent data storage")
	flag.StringVar(&logFile, "log-file", "my.log", "Path to log file")
	flag.StringVar(&listen, "listen", ":1234", "Listen address api interface")
	flag.StringVar(&peers, "peers", "", "Comma separated list of cluster peers")
	flag.BoolVar(&purgeOnExit, "purge-on-exit", false, "Purge all custom chains on exit")
	flag.Parse()

	if flagNotPassed("config-file") {
		configFile = getEnvString("CONFIG_FILE", "config.yml")
	}

	if len(configFile) > 0 {
		readConfig(configFile)
	}

	if flagNotPassed("listen") {
		s := getEnvString("LISTEN", config.Options.Listen)
		if len(s) > 0 {
			listen = s
		}
	}

	if flagNotPassed("data-path") {
		s := getEnvString("DATA_PATH", config.Options.DataPath)
		if len(s) > 0 {
			dataPath = s
		}
	}

	if flagNotPassed("log-file") {
		s := getEnvString("LOG_FILE", config.Options.LogFile)
		if len(s) > 0 {
			logFile = s
		}
	}

	if flagNotPassed("peers") {
		s := strings.Join(config.Options.Peers, ",")
		s = getEnvString("PEERS", s)
		if len(s) > 0 {
			peers = s
		}
	}

	if flagNotPassed("purge-on-exit") {
		purgeOnExit = config.Options.PurgeOnExit
	}

	// Configure logging
	// log.SetFlags(log.LstdFlags | log.Lshortfile)
	if len(logFile) > 0 {
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		w := io.MultiWriter(f, os.Stdout)
		log.SetOutput(w)
	}

	// Options
	log.Printf("registered configuration options")
	log.Printf("config-file = %s", configFile)
	log.Printf("data-path = %s", dataPath)
	log.Printf("log-file = %s", logFile)
	log.Printf("listen = %s", listen)
	log.Printf("peers = %s", peers)
	log.Printf("purge-on-exit = %v", purgeOnExit)

	// Create chains
	err := createChains()
	if err != nil {
		log.Fatal(err)
	}

	// Load static rules
	err = loadRules(dataPath, true)
	if err != nil {
		log.Fatal(err)
	}

	// Load registry
	err = registry.Init(dataPath)
	if err != nil {
		log.Fatal(err)
	}

	rs := registry.List()
	for _, id := range rs {
		s := registry.Get(id)
		err := iptables.CreateRules(s.Rule)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Start Hub
	go hub.Run()
	rs = strings.Split(peers, ",")
	for _, s := range rs {
		hub.Connect(s)
	}
	go hub.Reconnect()

	// HTTP Server
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/iptables/rules", RulesHandler)
	mux.HandleFunc("/api/v1/cluster", ClusterHandler)

	server := &http.Server{
		Addr:    listen,
		Handler: mux,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Handle signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT)

	s := <-sigs
	log.Printf("Caught %v signal, exiting...", s)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	onExit(purgeOnExit)
	cancel()
}
