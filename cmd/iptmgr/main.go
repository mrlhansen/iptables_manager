package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mrlhansen/iptables_manager/pkg/iptables"
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
	var cfgfile string
	var datadir string
	var listen string
	var logfile string
	var peer string
	var purge bool

	// Check for root or iptables permissions?

	flag.StringVar(&cfgfile, "config", "config.yml", "Path to configuration file")
	flag.StringVar(&datadir, "datadir", ".", "Path to persistent data storage")
	flag.StringVar(&logfile, "logfile", "my.log", "Path to log file")
	flag.StringVar(&listen, "listen", ":1234", "Listen address for web interface")
	flag.StringVar(&peer, "peer", "", "List of peers")
	flag.BoolVar(&purge, "purge-on-exit", false, "Purge all custom chains on exit")

	flag.Parse()
	readConfig(cfgfile)

	if flagNotPassed("listen") {
		s := getEnvString("LISTEN", config.Listen)
		if len(s) > 0 {
			listen = s
		}
	}

	if flagNotPassed("datadir") {
		s := getEnvString("DATADIR", config.DataDir)
		if len(s) > 0 {
			datadir = s
		}
	}

	if flagNotPassed("logfile") {
		s := getEnvString("LOGFILE", config.LogFile)
		if len(s) > 0 {
			logfile = s
		}
	}

	if flagNotPassed("purge-on-exit") {
		purge = config.Purge
	}

	// Configure logging
	// log.SetFlags(log.LstdFlags | log.Lshortfile)
	if len(logfile) > 0 {
		f, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
		if err != nil {
			log.Fatal("Failed to open log file")
		}
		w := io.MultiWriter(f, os.Stdout)
		log.SetOutput(w)
	}

	// Options
	log.Printf("registered configuration options")
	log.Printf("> config  = %s", cfgfile)
	log.Printf("> datadir = %s", datadir)
	log.Printf("> logfile = %s", logfile)
	log.Printf("> listen  = %s", listen)
	log.Printf("> purge-on-exit = %v", purge)

	// // Create chains
	// err := createChains()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// // Load static rules
	// err = loadRules(datadir, true)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// // Load registry
	// err = registry.Init(datadir)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// rs := registry.List()
	// for _, id := range rs {
	// 	s := registry.Get(id)
	// 	err := iptables.CreateRules(s)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// }

	// Start Hub
	go hub.run()

	if len(peer) > 0 {
		hub.connect(peer)
	}

	go hub.reconnect() // must go after all connection attempts

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

	onExit(purge)
	cancel()
}
