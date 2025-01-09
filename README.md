# iptables manager
This program is a very simple manager for loading custom iptables rules. When the program is started, it will create a set of custom chains, to which all subsequent rules are added. When the program stops, it will flush and delete these custom chains.

## Download
The manager is written in [Go](https://golang.org) and it can be downloaded and compiled using:
```bash
go get github.com/mrlhansen/iptables_manager
```

## Usage
The program accepts three arguments. The start and stop options can be added at the same time, which will result in a reload of all rules.
```
Usage of iptables_manager:
-confdir string
	path to the configuration directory (default "/etc/iptmgr")
-no-ipv4
    disable ipv4 rules
-no-ipv6
    disable ipv6 rules
-start
	start the manager
-stop
	stop the manager
```

## Operation
When the program starts it will initialize a new set of chains using the rules in `<confdir>/startX.rules` and when it stops it will delete these chains by reading the rules in `<confdir>/stopX.rules`. Editing these files might prevent the program from working properly, but they are kept as external files for flexibility. On startup, after initializing the new chains, it will proceed to read and apply the rules from all files in `<confdir>/rulesX.d` in ordered sequence.
