package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
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

func main() {
	var cfgfile string
	var datadir string
	var listen string
	var purge bool

	flag.StringVar(&cfgfile, "config", "config.yml", "Path to configuration file")
	flag.StringVar(&datadir, "datadir", ".", "Path to persistent data storage")
	flag.StringVar(&listen, "listen", ":1234", "Listen address for web interface")
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

	if flagNotPassed("purge-on-exit") {
		purge = config.Purge
	}

	log.Printf("registered configuration options")
	log.Printf("> config  = %s", cfgfile)
	log.Printf("> datadir = %s", datadir)
	log.Printf("> listen  = %s", listen)
	log.Printf("> purge-on-exit = %v", purge)

	// Create chains
	err := createChains()
	if err != nil {
		log.Fatal(err)
	}

	// Load static rules
	err = loadStaticRules(datadir, true)
	if err != nil {
		log.Fatal(err)
	}

	// Load registry
	err = registry.Init(datadir)
	if err != nil {
		log.Fatal(err)
	}

	rs := registry.List()
	for _, id := range rs {
		s := registry.Get(id)
		err := iptables.CreateRules(s)
		if err != nil {
			log.Fatal(err)
		}
	}

	// HTTP Server
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/iptables/rules", RulesHandler)
	mux.HandleFunc("/api/v1/iptables/chains", ChainsHandler)

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

	if purge {
		err := iptables.PurgeChains()
		if err != nil {
			log.Print(err)
		}
	}

	cancel()
}
