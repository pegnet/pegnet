package main

import (
	"os"
)

var commands map[string]*Command

func init() {
	commands = make(map[string]*Command)
	Help()
	NewAddress()
}

// Simple CLI executes commands to do things like convert FCT addresses to PNT addresses,
// show balances at addresses, convert assets from one asset type to another, and create
// and submit transactions
func main() {
	cmdn := "help"
	if len(os.Args) >= 2 {
		cmdn = os.Args[1]
	}
	if cmd, ok := commands[cmdn]; ok {
		cmd.Execute()
	}
}
