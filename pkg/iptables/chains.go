package iptables

import (
	"fmt"
	"strings"
)

type Chain struct {
	Name    string `json:"name"`
	Parent  string `json:"parent"`
	Insert  bool   `json:"insert"`
	Default bool   `json:"default"`
}

type Table struct {
	CustomChains []Chain
	DefaultChains map[string]string
}

var tables = map[string]Table{
	"filter": Table{
		CustomChains: []Chain{},
		DefaultChains: map[string]string{
			"input":   "INPUT",
			"output":  "OUTPUT",
			"forward": "FORWARD",
		},
	},
	"nat": Table{
		CustomChains: []Chain{},
		DefaultChains: map[string]string{
			"input":       "INPUT",
			"output":      "OUTPUT",
			"prerouting":  "PREROUTING",
			"postrouting": "POSTROUTING",
		},
	},
}

func FindChain(table string, name string) (int, error) {
	for n,k := range tables[table].CustomChains {
		if k.Name == name {
			return n, nil
		}
	}

	return 0, fmt.Errorf("unable to find chain (%s) in table (%s)", name, table)
}

func CreateChain(table string, chain *Chain) error {
	var err error

	name := chain.Name
	parent := chain.Parent

	if table == "filter" {
		err = checkPattern(pattern_filter_chain, name, "name")
		if err != nil {
			return err
		}
		err = checkPattern(pattern_filter_chain, parent, "parent")
		if err != nil {
			return err
		}
	} else {
		err = checkPattern(pattern_nat_chain, name, "name")
		if err != nil {
			return err
		}
		err = checkPattern(pattern_nat_chain, parent, "parent")
		if err != nil {
			return err
		}
	}

	_, n1, found := strings.Cut(name, "-")
	if !found {
		return fmt.Errorf("invalid chain name (%s)", name)
	}

	p1, p2, found := strings.Cut(parent, "-")
	if found {
		p1 = p2
	}

	if p1 != n1 {
		return fmt.Errorf("chain name (%s) does not match parent (%s)", name, parent)
	}

	if found {
		_, err := FindChain(table, chain.Parent)
		if err != nil {
			return err
		}
	} else {
		parent = strings.ToUpper(parent)
	}

	if chain.Default {
		tables[table].DefaultChains[n1] = name
	}

	err = iptables_create_chain(table, name)
	if err != nil {
		return err
	}

	err = iptables_link_chains(table, name, parent)
	if err != nil {
		return err
	}

	if entry, ok := tables[table]; ok {
		entry.CustomChains = append(entry.CustomChains, *chain)
		tables[table] = entry
	}

	return nil
}
