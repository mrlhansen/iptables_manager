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

	str := "iptables -w " + strings.Join(arg[:], " ")
	cmd := exec.Command("bash", "-c", str)
	log.Print(str)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	return (err == nil)
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

func iptables_link_chain(table string, chain string, parent string, insert bool) error {
	// Check if chains are already linked
	cmd := []string{"-t", table, "-C", parent, "-j", chain}
	ok := iptables_run_command(cmd)
	if ok {
		return nil
	}

	// Append or insert
	opt := "-A"
	if insert {
		opt = "-I"
	}

	// Link chains
	cmd = []string{"-t", table, opt, parent, "-j", chain}
	ok = iptables_run_command(cmd)
	if !ok {
		return fmt.Errorf("failed to link chain (%s) to parent chain (%s) in table (%s)", chain, parent, table)
	}

	return nil
}

func iptables_delete_chain(table string, chain string) error {
	// Check if chain exists
	cmd := []string{"-t", table, "-L", chain}
	ok := iptables_run_command(cmd)
	if !ok {
		return nil
	}

	// Flush chain
	cmd = []string{"-t", table, "-F", chain}
	ok = iptables_run_command(cmd)
	if !ok {
		return fmt.Errorf("failed to flush chain (%s) in table (%s)", chain, table)
	}

	// Delete chain
	cmd = []string{"-t", table, "-X", chain}
	ok = iptables_run_command(cmd)
	if !ok {
		return fmt.Errorf("failed to delete chain (%s) in table (%s)", chain, table)
	}

	return nil
}

func iptables_unlink_chain(table string, chain string, parent string) error {
	// Check if chains are linked
	cmd := []string{"-t", table, "-C", parent, "-j", chain}
	ok := iptables_run_command(cmd)
	if !ok {
		return nil
	}

	// Unlink chains
	cmd = []string{"-t", table, "-D", parent, "-j", chain}
	ok = iptables_run_command(cmd)
	if !ok {
		return fmt.Errorf("failed to unlink chain (%s) from parent chain (%s) in table (%s)", chain, parent, table)
	}

	return nil
}

func iptables_create_rule(rule string) error {
	// Check if the rule exists
	chk := strings.Replace(rule, "-A ", "-C ", 1)
	chk = strings.Replace(chk, "-I ", "-C ", 1)
	cmd := []string{chk}
	ok := iptables_run_command(cmd)
	if ok {
		return nil
	}

	// Add rule
	cmd = []string{rule}
	ok = iptables_run_command(cmd)
	if !ok {
		return fmt.Errorf("failed to create rule: %s", rule)
	}

	return nil
}

func iptables_delete_rule(rule string) error {
	// Check if the rule exists
	chk := strings.Replace(rule, "-A ", "-C ", 1)
	chk = strings.Replace(chk, "-I ", "-C ", 1)
	cmd := []string{chk}
	ok := iptables_run_command(cmd)
	if !ok {
		return nil
	}

	// Delete rule
	chk = strings.Replace(chk, "-C ", "-D ", 1)
	cmd = []string{chk}
	ok = iptables_run_command(cmd)
	if !ok {
		return fmt.Errorf("failed to delete rule: %s", rule)
	}

	return nil
}
