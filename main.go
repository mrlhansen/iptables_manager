package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
)

func lsdir(path string) []string {
	var files []string

	list, err := os.ReadDir(path)
	if err != nil {
		log.Fatal("unable to read directory ", path)
	}

	for _, file := range list {
		if !file.IsDir() {
			files = append(files, path+"/"+file.Name())
		}
	}

	sort.Strings(files)
	return files
}

func read_rules(filename string, replace bool) []string {
	var lines []string
	var text string

	content, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal("unable to read file ", filename)
	} else {
		log.Print("reading file ", filename)
	}
	text = string(content)

	if replace {
		text = strings.ReplaceAll(text, "INPUT", "iptmgr-input")
		text = strings.ReplaceAll(text, "OUTPUT", "iptmgr-output")
		text = strings.ReplaceAll(text, "FORWARD", "iptmgr-forward")
		text = strings.ReplaceAll(text, "PREROUTING", "iptmgr-prerouting")
		text = strings.ReplaceAll(text, "POSTROUTING", "iptmgr-postrouting")
	}

	lines = strings.Split(text, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 0 && line[0] == '#' {
			line = ""
		}
		lines[i] = line
	}

	return lines
}

func apply_rules(filename string, rules []string, iptcmd string) bool {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	for i, rule := range rules {
		if len(rule) == 0 {
			continue
		}

		cmd := exec.Command("bash", "-c", fmt.Sprintf("%s -w %s", iptcmd, rule))
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()

		if err != nil {
			s := strings.Split(stderr.String(), "\n")
			log.Printf("%s:%d %s\n", filename, i, s[0])
			return false
		}
	}

	return true
}

func manager_load_rules(path, iptcmd string) bool {
	files := lsdir(path)
	for _, file := range files {
		rules := read_rules(file, true)
		ok := apply_rules(file, rules, iptcmd)
		if !ok {
			return false
		}
	}
	return true
}

func manager_load_chains(file, iptcmd string) bool {
	rules := read_rules(file, false)
	ok := apply_rules(file, rules, iptcmd)
	return ok
}

func manager_stop(path string, ipv4, ipv6 bool) {
	if ipv4 {
		manager_load_chains(path+"/stop4.rules", "iptables")
	}
	if ipv6 {
		manager_load_chains(path+"/stop6.rules", "ip6tables")
	}
}

func manager_start(path string, ipv4, ipv6 bool) bool {
	if ipv4 {
		ok := manager_load_chains(path+"/start4.rules", "iptables")
		if !ok {
			return false
		}
	}
	if ipv6 {
		ok := manager_load_chains(path+"/start6.rules", "ip6tables")
		if !ok {
			return false
		}
	}
	return true
}

func manager_rules(path string, ipv4, ipv6 bool) bool {
	if ipv4 {
		ok := manager_load_rules(path+"/rules4.d", "iptables")
		if !ok {
			return false
		}
	}
	if ipv6 {
		ok := manager_load_rules(path+"/rules6.d", "ip6tables")
		if !ok {
			return false
		}
	}
	return true
}

func main() {
	var (
		path  string
		start bool
		stop  bool
		ipv4  bool
		ipv6  bool
	)

	// Parse flags
	flag.StringVar(&path, "confdir", "/etc/iptmgr", "path to the configuration directory")
	flag.BoolVar(&start, "start", false, "start the manager")
	flag.BoolVar(&stop, "stop", false, "stop the manager")
	flag.BoolVar(&ipv4, "no-ipv4", false, "disable ipv4 rules")
	flag.BoolVar(&ipv6, "no-ipv6", false, "disable ipv6 rules")
	flag.Parse()

	// Toogle
	ipv4 = !ipv4
	ipv6 = !ipv6

	// Remove chains
	if stop {
		manager_stop(path, ipv4, ipv6)
	}

	// Create chains
	if start {
		ok := manager_start(path, ipv4, ipv6)
		if !ok {
			os.Exit(1)
		}
	}

	// Load rules
	if start {
		ok := manager_rules(path, ipv4, ipv6)
		if !ok {
			manager_stop(path, ipv4, ipv6)
		}
	}
}
