package iptables

import (
	"fmt"
	"strings"

	"github.com/mrlhansen/iptables_manager/pkg/registry"
)

type Rule struct {
	Table                string `json:"table"`                // -t
	Chain                string `json:"chain"`                // -A
	Action               string `json:"action"`               // -j
	Protocol             string `json:"protocol"`             // -p
	SourceInterface      string `json:"sourceInterface"`      // -i
	DestinationInterface string `json:"destinationInterface"` // -o
	SourceSubnet         string `json:"sourceSubnet"`         // -s
	DestinationSubnet    string `json:"destinationSubnet"`    // -d
	SourcePorts          string `json:"sourcePorts"`          // --sport or --sports
	DestinationPorts     string `json:"destinationPorts"`     // --dport or --dports
	NatDestination       string `json:"natDestination"`       // --to-destination or --to-source
}

func BuildRule(m *Rule, comment string) (string, error) {
	var cmd strings.Builder
	var err error
	var nat bool
	var ports bool
	var multiport bool

	// Table
	table := strings.ToLower(m.Table)
	err = checkPattern("nat|filter", table, "table")
	if err != nil {
		return "", err
	}
	cmd.WriteString("-t " + table)

	// Chain
	chain := strings.ToLower(m.Chain)
	err = ValidateChain(table, chain, true, true)
	if err != nil {
		return "", err
	}
	chain = DefaultChain(table, chain)
	cmd.WriteString(" -A " + chain)

	// SourceInterface
	value := m.SourceInterface
	if len(value) > 0 {
		err = checkPattern(pattern_interface, value, "sourceInterface")
		if err != nil {
			return "", err
		}
		cmd.WriteString(" -i " + value)
	}

	// DestinationInterface
	value = m.DestinationInterface
	if len(value) > 0 {
		err = checkPattern(pattern_interface, value, "destinationInterface")
		if err != nil {
			return "", err
		}
		cmd.WriteString(" -o " + value)
	}

	// SourceSubnet
	value = m.SourceSubnet
	if len(value) > 0 {
		err = checkPattern(pattern_subnet, value, "sourceSubnet")
		if err != nil {
			return "", err
		}
		cmd.WriteString(" -s " + value)
	}

	// DestinationSubnet
	value = m.DestinationSubnet
	if len(value) > 0 {
		err = checkPattern(pattern_subnet, value, "destinationSubnet")
		if err != nil {
			return "", err
		}
		cmd.WriteString(" -d " + value)
	}

	// SourcePorts
	sports := m.SourcePorts
	if len(sports) > 0 {
		err = checkPattern(pattern_ports, sports, "sourcePorts")
		if err != nil {
			return "", err
		}
		if strings.Contains(sports, ",") {
			multiport = true
			sports = " --sports " + sports
		} else {
			sports = " --sport " + sports
		}
		ports = true
	}

	// DestinationPorts
	dports := m.DestinationPorts
	if len(dports) > 0 {
		err = checkPattern(pattern_ports, dports, "destinationPorts")
		if err != nil {
			return "", err
		}
		if strings.Contains(dports, ",") {
			multiport = true
			dports = " --dports " + dports
		} else {
			dports = " --dport " + dports
		}
		ports = true
	}

	// Protocol
	value = strings.ToLower(m.Protocol)
	if len(value) > 0 {
		err = checkPattern("tcp|udp", value, "protocol")
		if err != nil {
			return "", err
		}

		cmd.WriteString(" -p " + value)
		if multiport {
			cmd.WriteString(" -m multiport")
		}
		if len(dports) > 0 {
			cmd.WriteString(dports)
		}
		if len(sports) > 0 {
			cmd.WriteString(sports)
		}
	} else {
		if ports {
			return "", fmt.Errorf("protocol must be specified when using port selectors")
		}
	}

	// Comment
	if len(comment) > 0 {
		cmd.WriteString(` -m comment --comment "` + comment + `"`)
	}

	// Action
	action := strings.ToLower(m.Action)
	err = checkPattern("dnat|snat", action, "action")
	if err == nil {
		nat = true
	}
	if chain == "prerouting" {
		err = checkPattern("dnat|accept|drop", action, "action")
		if err != nil {
			return "", err
		}
	} else if chain == "postrouting" {
		err = checkPattern("snat|masquerade|accept|drop", action, "action")
		if err != nil {
			return "", err
		}
	} else {
		err = checkPattern("accept|drop", action, "action")
		if err != nil {
			return "", err
		}
	}
	cmd.WriteString(" -j " + strings.ToUpper(action))

	// NatDestination
	value = m.NatDestination
	if len(value) > 0 {
		if !nat {
			return "", fmt.Errorf("action (%s) does not support use of nat destination", action)
		}

		err = checkPattern(pattern_destination, value, "natDestination")
		if err != nil {
			return "", err
		}

		if action == "snat" {
			cmd.WriteString(" --to-source " + value)
		} else {
			cmd.WriteString(" --to-destination " + value)
		}
	} else {
		if nat {
			return "", fmt.Errorf("action (%s) requires use of nat destination", action)
		}
	}

	return cmd.String(), nil
}

func DeleteRules(s string) error {
	rs := strings.Split(s, "\n")

	for _, rule := range rs {
		rule = strings.TrimSpace(rule)
		if len(rule) == 0 {
			continue
		}
		if rule[0] == '#' {
			continue
		}

		err := iptables_delete_rule(rule)
		if err != nil {
			// do we want to continue and try the rest?
			return err
		}
	}

	return nil
}

func CreateRules(s string) error {
	rs := strings.Split(s, "\n")

	for _, rule := range rs {
		rule = strings.TrimSpace(rule)
		if len(rule) == 0 {
			continue
		}
		if rule[0] == '#' {
			continue
		}

		err := iptables_create_rule(rule)
		if err != nil {
			// roll back and fail - we can call DeleteRules on the same input
			return err
		}
	}

	return nil
}

func CreateRuleSet(rules []Rule) (string, error) {
	var rs []string = []string{}

	mu.Lock()
	defer mu.Unlock()

	id, _ := registry.GenerateName()

	for n := range rules {
		s, err := BuildRule(&rules[n], id)
		if err != nil {
			return "", err
		}
		rs = append(rs, s)
	}

	s := strings.Join(rs, "\n")
	err := CreateRules(s)
	if err != nil {
		return "", err
	}

	err = registry.Append(id, s)
	if err != nil {
		// roll back and fail - we can call DeleteRules on the same input
		return "", err
	}

	return id, nil
}

func DeleteRuleSet(id string) error {
	s := registry.Get(id)
	if s == "" {
		return nil // or not found
	}

	err := DeleteRules(s)
	if err != nil {
		return err
	}

	err = registry.Delete(id)
	if err != nil {
		return err
	}

	return nil
}
