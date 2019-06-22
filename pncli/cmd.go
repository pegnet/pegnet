package main

type Command struct {
	Name      string
	ShortHelp string
	LongHelp  string
	Execute   func()
}
