package main

type Command struct {
	Name      string // Name of the command
	Parms     string // Parameters for the command
	ShortHelp string // Short one line help
	LongHelp  string // Long help for the command
	Execute   func() // Actually execute the command with this function
}
