package main

import (
	"fmt"
	"io/ioutil"
)

func main() {
	data, err := ioutil.ReadFile("apikey.dat")
	fmt.Printf("%0x %v", data, err)
}
