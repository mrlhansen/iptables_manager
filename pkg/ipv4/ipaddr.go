package ipv4

import (
	"fmt"
	"log"
	"sync"
)

type Address struct {
	Address string `json:"cidr"`
	Netdev  string `json:"netdev"`
}

var state bool = true
var mu sync.Mutex

func toggleState(st bool) {
	mu.Lock()
	defer mu.Unlock()

	if st {
		log.Print("We are now active")
	} else {
		log.Print("We are now backup")
	}
}

func SetState(st bool) {
	if state != st {
		state = st
		go toggleState(st)
	}
}

func CreateAddress(a *Address) error {
	ok := ipv4_addr_validate(a.Address)
	if !ok {
		return fmt.Errorf("ipv4 address (%s) is invalid", a.Address)
	}

	ok = ipv4_netdev_exists(a.Netdev)
	if !ok {
		return fmt.Errorf("netdev (%s) does not exist", a.Netdev)
	}

	ok = ipv4_addr_ping(a.Address)
	if ok {
		return fmt.Errorf("address (%s) is already reachable", a.Address)
	}

	// Add to "registry" or whatever

	if state {
		ok = ipv4_addr_add(a.Netdev, a.Address)
		if !ok {
			return fmt.Errorf("failed to add ipv4 address (%s) to netdev (%s)", a.Address, a.Netdev)
		}
	}

	return nil
}
