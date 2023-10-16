package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/mrlhansen/iptables_manager/pkg/iptables"
)

func sanitizeRules(s *string) error {
	nre := regexp.MustCompile(`-t\s+nat`)
	cre := regexp.MustCompile(`-[AI]\s+([\w\-]+)`)

	raw := strings.Split(*s, "\n")
	out := []string{}

	for _, rule := range raw {
		rule = strings.TrimSpace(rule)
		if len(rule) == 0 {
			continue
		}
		if rule[0] == '#' {
			continue
		}

		table := "filter"
		if ok := nre.MatchString(rule); ok {
			table = "nat"
		}

		match := cre.FindStringSubmatch(rule)
		if len(match) != 2 {
			return fmt.Errorf("unable to determine chain: %s", rule)
		}

		rawchain := match[1]
		chain := strings.ToLower(rawchain)

		err := iptables.ValidateChain(table, chain, true, true)
		if err != nil {
			return err
		}

		chain = iptables.DefaultChain(table, chain)
		rule = strings.Replace(rule, rawchain, chain, 1)
		out = append(out, rule)
	}

	*s = strings.Join(out, "\n")
	return nil
}

func loadStaticRules(basepath string, defchain bool) error {
	var files []string
	var rs []string

	path := basepath + "/rules"
	list, err := os.ReadDir(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	for _, file := range list {
		if !file.IsDir() {
			files = append(files, file.Name())
		}
	}

	sort.Strings(files)

	for _, fn := range files {
		path := basepath + "/rules/" + fn
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		s := string(data)
		if err := sanitizeRules(&s); err != nil {
			log.Print(err)
			return nil
		}

		rs = append(rs, s)
	}

	for _, s := range rs {
		err := iptables.CreateRules(s)
		if err != nil {
			return err
		}
	}

	return nil
}
