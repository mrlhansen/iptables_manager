package main

import (
	"flag"
	"bytes"
	"io/ioutil"
	"os/exec"
	"strings"
	"sort"
	"log"
)

func lsdir(path string) []string {
	var files []string

	list, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal("unable to read directory ", path)
	}

	for _,file := range list {
		if file.IsDir() == false {
			files = append(files, path+"/"+file.Name())
		}
	}

	sort.Strings(files)
	return files
}

func read_rules(filename string, replace bool) []string {
	var lines []string
	var text string

	content, err := ioutil.ReadFile(filename)
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
	for i,line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 0 && line[0] == '#' {
			line = ""
		}
		lines[i] = line
	}

	return lines
}

func apply_rules(filename string, rules []string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	for i,rule := range rules {
		if len(rule) == 0 {
			continue
		}

		cmd := exec.Command("bash", "-c", "iptables " + rule)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()

		if err != nil {
			s := strings.Split(stderr.String(), "\n")
			log.Fatalf("%s:%d %s\n", filename, i, s[0])
		}
	}
}

func manager_start(path string) {
	file := path + "/start.rules"
	files := lsdir(path + "/rules.d")
	rules := read_rules(file, false)
	apply_rules(file, rules)

	for _,file = range files {
		rules = read_rules(file, true)
		apply_rules(file, rules)
	}
}

func manager_stop(path string) {
	file := path + "/stop.rules"
	rules := read_rules(file, false)
	apply_rules(file, rules)
}

func main() {
	var path string
	var start *bool
	var stop *bool

	flag.StringVar(&path, "confdir", "/etc/iptmgr", "path to the configuration directory")
	start = flag.Bool("start", false, "start the manager")
	stop = flag.Bool("stop", false, "stop the manager")
	flag.Parse()

	if *stop {
		manager_stop(path)
	}

	if *start {
		manager_start(path)
	}
}
