package main

import (
	"log"
	"os"

	"github.com/mrlhansen/iptables_manager/pkg/iptables"
	"gopkg.in/yaml.v2"
)

type ConfigChains struct {
	Filter []iptables.Chain `yaml:"filter"`
	Nat    []iptables.Chain `yaml:"nat"`
}

type Config struct {
	DataDir string        `yaml:"datadir"`
	Chains  *ConfigChains `yaml:"chains"`
}

var config = Config{}

func readConfig(fn string) {
	fp, err := os.Open(fn)
	if err != nil {
		log.Fatalf("error reading configuration file %s: %s", fn, err)
	}

	err = yaml.NewDecoder(fp).Decode(&config)
	fp.Close()
	if err != nil {
		log.Fatalf("error parsing configuration file %s: %s", fn, err)
	}
}

func createChains() {
	if config.Chains == nil {
		return
	}

	for _, chain := range config.Chains.Filter {
		err := iptables.CreateChain("filter", &chain)
		if err != nil {
			log.Fatal(err)
		}
	}

	for _, chain := range config.Chains.Nat {
		err := iptables.CreateChain("nat", &chain)
		if err != nil {
			log.Fatal(err)
		}
	}
}
