package ipv4

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

func ipv4_run_command(arg []string, o *string, e *string) bool {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	str := strings.Join(arg[:], " ")
	cmd := exec.Command("bash", "-c", str)
	log.Print(str)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if o != nil {
		*o = stdout.String()
	}

	if e != nil {
		*e = stderr.String()
	}

	return (err == nil)
}

func ipv4_addr_validate(addr string) bool {
	ok, _ := regexp.MatchString(`^(\d{1,3}\.){3}\d{1,3}/\d{1,2}$`, addr)
	return ok
}

func ipv4_addr_exists(netdev string, addr string) bool {
	var stdout string

	cmd := []string{"ip", "-4", "-br", "-o", "addr"}
	ipv4_run_command(cmd, &stdout, nil)

	pattern := fmt.Sprintf(`(^%s)(\s.*\s)(%s)(\s|$)`, regexp.QuoteMeta(netdev), regexp.QuoteMeta(addr))
	re := regexp.MustCompile(pattern)

	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		ok := re.MatchString(line)
		if ok {
			return true
		}
	}

	return false
}

func ipv4_netdev_exists(netdev string) bool {
	var stdout string

	cmd := []string{"ip", "-br", "-o", "link"}
	ipv4_run_command(cmd, &stdout, nil)

	pattern := fmt.Sprintf(`(^%s)(\s)`, regexp.QuoteMeta(netdev))
	re := regexp.MustCompile(pattern)

	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		ok := re.MatchString(line)
		if ok {
			return true
		}
	}

	return false
}

func ipv4_addr_add(netdev string, addr string) bool {
	ok := ipv4_addr_exists(netdev, addr)
	if ok {
		return true
	}

	cmd := []string{"ip", "addr", "add", addr, "dev", netdev}
	ok = ipv4_run_command(cmd, nil, nil)
	return ok
}

func ipv4_addr_del(netdev string, addr string) bool {
	ok := ipv4_addr_exists(netdev, addr)
	if !ok {
		return true
	}

	cmd := []string{"ip", "addr", "del", addr, "dev", netdev}
	ok = ipv4_run_command(cmd, nil, nil)
	return ok
}

// this wont work - cant ping an addresses if we do not already have a route for that subnet. Need an ARP ping.
// https://pkg.go.dev/github.com/j-keck/arping
func ipv4_addr_ping(addr string) bool {
	addr, _, _ = strings.Cut(addr, "/")
	cmd := []string{"ping", "-q", "-W", "0.25", "-c", "1", addr}
	ok := ipv4_run_command(cmd, nil, nil)
	return ok
}
