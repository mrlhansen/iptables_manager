package iptables

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func iptables_run_command(arg []string) bool {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	str := strings.Join(arg[:], " ")
	cmd := exec.Command("bash", "-c", "iptables -w "+str)
	log.Print("iptables -w " + str)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		return false
	}

	return true
}

func iptables_create_chain(table string, chain string) error {
	// Check if chain already exists
	cmd := []string{"-t", table, "-L", chain}
	ok := iptables_run_command(cmd)
	if ok {
		return nil
	}

	// Create chain
	cmd = []string{"-t", table, "-N", chain}
	ok = iptables_run_command(cmd)
	if !ok {
		return fmt.Errorf("failed to create chain (%s) in table (%s)", chain, table)
	}

	return nil
}

func iptables_link_chains(table string, chain string, parent string) error {
	// Check if chains are already linked
	cmd := []string{"-t", table, "-C", parent, "-j", chain}
	ok := iptables_run_command(cmd)
	if ok {
		return nil
	}

	// Link chains
	cmd = []string{"-t", table, "-I", parent, "-j", chain}
	ok = iptables_run_command(cmd)
	if !ok {
		return fmt.Errorf("failed to link chain (%s) to parent chain (%s) in table (%s)", chain, parent, table)
	}

	return nil
}
