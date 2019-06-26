package main

import (
	"fmt"
	"github.com/FactomProject/factom"
	"github.com/pegnet/pegnet/common"
	"sort"
	"strings"
)

var _ = func() (n int) {
	Init()
	me := Command{
		Name:      "newaddress",
		ShortHelp: "creates a new FCT address in your wallet, and provides the list of all asset addresses.",
		LongHelp: "" +
			"Creates a new FCT address and puts the private key for that address in your wallet. All\n" +
			"PegNet assets are secured by the same private key, and this command provides you the\n" +
			"human/wallet addresses to use to access those assets." +
			"\n" +
			"Usage:\n" +
			"\n" +
			"pncli newaddress\n" +
			"",
		Execute: func() {
			fa, err := factom.GenerateFactoidAddress()
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Print(fa.String(), "\n\n")
			assets, err := common.ConvertUserFctToUserPegNetAssets(fa.String())
			if err != nil {
				fmt.Println(err)
				return
			}
			sort.Strings(assets)
			for _, s := range assets {
				if strings.Contains(s, "PNT_") {
					fmt.Println("  *", s[1:4], " ", s)
				} else {
					fmt.Println("   ", s[1:4], " ", s)
				}
			}
		},
	}
	commands[strings.ToLower(me.Name)] = &me
	return
}()
