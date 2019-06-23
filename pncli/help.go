package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

var _ = func() (n int) {
	Init()
	me := Command{
		Name:      "help",
		ShortHelp: "Returns a list of commands, or pncli help <cmd> returns long help on a command.",
		LongHelp:  "Returns a list of commands, or pncli help <cmd> returns long help on a command.",
		Execute: func() {
			// If we only have "help", print a list of short help for all commands
			if len(os.Args) > 2 {
				cmd := commands[os.Args[2]]
				if cmd != nil {
					fmt.Println("\n", cmd.Name, " ", cmd.Parms)
					fmt.Println("\n", cmd.LongHelp)
					return
				}
			}
			var clist []*Command
			for _, cmd := range commands {
				clist = append(clist, cmd)
			}
			sort.Slice(clist, func(i, j int) bool {
				return clist[i].Name < clist[j].Name
			})
			for _, cmd := range clist {
				fmt.Printf("%20s -- %s\n", cmd.Name, cmd.ShortHelp)
			}
		},
	}
	commands[strings.ToLower(me.Name)] = &me
	return
}()
