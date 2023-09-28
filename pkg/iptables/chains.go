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
	CustomChains  []Chain
	DefaultChains map[string]string
}

var tables = map[string]Table{
	"filter": {
		CustomChains: []Chain{},
		DefaultChains: map[string]string{
			"input":   "INPUT",
			"output":  "OUTPUT",
			"forward": "FORWARD",
		},
	},
	"nat": {
		CustomChains: []Chain{},
		DefaultChains: map[string]string{
			"input":       "INPUT",
			"output":      "OUTPUT",
			"prerouting":  "PREROUTING",
			"postrouting": "POSTROUTING",
		},
	},
}

func ValidateChain(table string, chain string, exists bool, allowroot bool) error {
	if table == "filter" {
		err := checkPattern(pattern_filter_chain, chain, "chain")
		if err != nil {
			return err
		}
	} else {
		err := checkPattern(pattern_nat_chain, chain, "chain")
		if err != nil {
			return err
		}
	}

	if strings.Contains(chain, "-") {
		_, err := FindChain(table, chain)
		if exists {
			if err != nil {
				return err
			}
			return nil
		} else {
			exists = (err == nil)
		}
	} else {
		if !allowroot {
			return fmt.Errorf("chain (%s) is a root chain in table (%s)", chain, table)
		}
		exists = !exists
	}

	if exists {
		return fmt.Errorf("chain (%s) already exists in table (%s)", chain, table)
	}

	return nil
}

func FindChain(table string, name string) (int, error) {
	for n, k := range tables[table].CustomChains {
		if k.Name == name {
			return n, nil
		}
	}

	return 0, fmt.Errorf("unable to find chain (%s) in table (%s)", name, table)
}

func DefaultChain(table string, chain string) string {
	if strings.Contains(chain, "-") {
		return chain
	}
	return tables[table].DefaultChains[chain]
}

func CreateChain(table string, chain *Chain) error {
	var err error

	mu.Lock()
	defer mu.Unlock()

	name := chain.Name
	parent := chain.Parent

	err = ValidateChain(table, parent, true, true)
	if err != nil {
		return err
	}

	err = ValidateChain(table, name, false, false)
	if err != nil {
		return err
	}

	_, n1, _ := strings.Cut(name, "-")
	p1, p2, found := strings.Cut(parent, "-")
	if found {
		p1 = p2
	}

	if p1 != n1 {
		return fmt.Errorf("chain name (%s) does not match parent (%s)", name, parent)
	}

	if !found {
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
