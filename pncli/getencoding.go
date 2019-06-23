package main

import (
	"fmt"
	"github.com/pegnet/pegnet/support"
	"os"
	"sort"
	"strings"
)

var _ = func() (n int) {
	Init()
	me := Command{
		Name:      "getencoding",
		Parms:     "<fct address> [encoding]",
		ShortHelp: "Takes a FCT address and returns an asset encoding (or all encodings) for that FCT address",
		LongHelp: "" +
			"All Pegnet assets are controlled by the same private key as a FCT address. You can specify\n" +
			"an asset, and this command will give you the encoding for that asset.  If you specify 'all',\n" +
			"or you don't specify an asset, you will get all assets." +
			"\n" +
			"Usage:\n" +
			"\n" +
			"    pncli getencoding FA2RwVjKe4Jrr7M7E62fZi8mFYqEAoQppmpEDXqAumGkiropSAbk usd\n" +
			"\n" +
			"or\n" +
			"\n" +
			"    pncli getencoding FA2RwVjKe4Jrr7M7E62fZi8mFYqEAoQppmpEDXqAumGkiropSAbk all\n",

		Execute: func() {
			asset := "all"
			switch len(os.Args) {
			case 2:
				fmt.Println("Must provide a FCT public key")
				return
			case 3:
			default:
				asset = strings.ToLower(os.Args[2])
			}
			assets, err := support.ConvertUserFctToUserPegNetAssets(os.Args[2])
			if err != nil {
				fmt.Println("Must provide a valid FCT public key")
				return
			}
			sort.Strings(assets)
			for _, s := range assets {
				if asset == "all" || asset == strings.ToLower(s[1:4]) {
					if strings.Contains(s, "PNT_") {
						fmt.Println("  *", s[1:4], " ", s)
					} else {
						fmt.Println("   ", s[1:4], " ", s)
					}
				}
			}
		},
	}
	commands[strings.ToLower(me.Name)] = &me
	return
}()
