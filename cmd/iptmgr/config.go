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
	Listen  string        `yaml:"listen"`
	Purge   bool          `yaml:"purge-on-exit"`
	DataDir string        `yaml:"datadir"`
	LogFile string        `yaml:"logfile"`
	Chains  *ConfigChains `yaml:"chains"`
}

var config = Config{}

func getEnvString(env string, def string) string {
	value := os.Getenv(env)
	if len(value) == 0 {
		return def
	}
	return value
}

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
