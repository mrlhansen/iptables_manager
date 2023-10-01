package iptables

import (
	"fmt"
	"regexp"
	"sync"
)

var mu sync.Mutex

const (
	pattern_subnet       = `(\d{1,3}\.){3}\d{1,3}/\d{1,2}`
	pattern_destination  = `(\d{1,3}\.){3}\d{1,3}(:\d{1,4})?`
	pattern_interface    = `[a-z0-9]+(\.(\d{1,4}|0x[0-9a-f]{1,4}))?`
	pattern_ports        = `(\d{1,5}(:\d{1,5})?,)*?(\d{1,5}(:\d{1,5})?)`
	pattern_nat_chain    = `(\w+\-)?(input|output|prerouting|postrouting)`
	pattern_filter_chain = `(\w+\-)?(input|output|forward)`
)

func checkPattern(pattern, value, name string) error {
	ok, _ := regexp.MatchString("^("+pattern+")$", value)
	if !ok {
		return fmt.Errorf("invalid value (%s) for argument (%s)", value, name)
	}
	return nil
}
