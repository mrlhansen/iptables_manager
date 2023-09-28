package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/mrlhansen/iptables_manager/pkg/iptables"
	"github.com/mrlhansen/iptables_manager/pkg/registry"
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

func apply_rules(filename string, rules []string) bool {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	for i, rule := range rules {
		if len(rule) == 0 {
			continue
		}

		cmd := exec.Command("bash", "-c", "iptables -w "+rule)
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

func manager_start(path string) {
	file := path + "/start.rules"
	files := lsdir(path + "/rules.d")
	rules := read_rules(file, false)

	ok := apply_rules(file, rules)
	if !ok {
		os.Exit(1)
	}

	for _, file = range files {
		rules = read_rules(file, true)
		ok := apply_rules(file, rules)
		if !ok {
			manager_stop(path)
			os.Exit(1)
		}
	}
}

func manager_stop(path string) {
	file := path + "/stop.rules"
	rules := read_rules(file, false)

	ok := apply_rules(file, rules)
	if !ok {
		os.Exit(1)
	}
}

func main() {
	var path string
	var start *bool
	var stop *bool

	flag.StringVar(&path, "confdir", "/etc/iptmgr", "path to the configuration directory")
	start = flag.Bool("start", false, "start the manager")
	stop = flag.Bool("stop", false, "stop the manager")
	flag.Parse()

	h := iptables.Chain{
		Name:    "heej-forward",
		Parent:  "forward",
		Insert:  true,
		Default: false,
	}

	err := iptables.CreateChain("filter", &h)
	if err != nil {
		log.Print(err)
	}

	k := iptables.Rule{
		Table:           "filter",
		Chain:           "forward",
		Action:          "accept",
		SourceInterface: "bond0",
	}

	str, err := iptables.BuildRule(&k)
	if err != nil {
		log.Print(err)
	} else {
		log.Print(str)
	}

	registry.Init(".")
	registry.New(str)

	if *stop {
		manager_stop(path)
	}

	if *start {
		manager_start(path)
	}
}
