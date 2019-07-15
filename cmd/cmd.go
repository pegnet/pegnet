package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/FactomProject/factom"
	"github.com/pegnet/pegnet/common"
	"github.com/spf13/cobra"
)

func init() {
	// Add commands to the root cmd
	rootCmd.AddCommand(getEncoding)
	rootCmd.AddCommand(newAddress)
	rootCmd.AddCommand(burn)
}

var getEncoding = &cobra.Command{
	Use:     "getencoding <fct address> [encoding]",
	Short:   "Takes a FCT address and returns an asset encoding (or all encodings) for that FCT address",
	Example: "pegnet getencoding FA2RwVjKe4Jrr7M7E62fZi8mFYqEAoQppmpEDXqAumGkiropSAbk usd\npncli getencoding FA2RwVjKe4Jrr7M7E62fZi8mFYqEAoQppmpEDXqAumGkiropSAbk all",
	// TODO: Verify this functionality.
	ValidArgs: ValidOwnedFCTAddresses(),

	Long: "" +
		"All Pegnet assets are controlled by the same private key as a FCT address. You can specify\n" +
		"an asset, and this command will give you the encoding for that asset.  If you specify 'all',\n" +
		"or you don't specify an asset, you will get all assets.",
	// TODO: Check the encoding is a valid option
	Args: CombineCobraArgs(cobra.RangeArgs(1, 2), CustomArgOrderValidationBuilder(false, ArgValidatorFCTAddress, ArgValidatorAssetAndAll)),
	Run: func(cmd *cobra.Command, args []string) {
		asset := "all"
		if len(args) == 2 {
			asset = strings.ToLower(args[1])
		}
		assets, err := common.ConvertUserFctToUserPegNetAssets(os.Args[2])
		if err != nil {
			// TODO: Verify this error message?
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

// addGetEncodingCommands adds commands so the autocomplete can fill in the second param
func addGetEncodingCommands() {
	for _, ass := range append([]string{"all"}, common.AssetNames...) {
		getEncoding.AddCommand(&cobra.Command{Use: strings.ToLower(ass), Run: func(cmd *cobra.Command, args []string) {}})
	}
}

var newAddress = &cobra.Command{
	Use:   "newaddress",
	Short: "creates a new FCT address in your wallet, and provides the list of all asset addresses.",
	Long: "Creates a new FCT address and puts the private key for that address in your wallet. All " +
		"PegNet assets are secured by the same private key, and this command provides you the " +
		"human/wallet addresses to use to access those assets",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
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


var burn = &cobra.Command{
	Use:"burn",
	Short:"Burns the specied amount of FCT into PNT",
	Args: CombineCobraArgs(CustomArgOrderValidationBuilder(true, ArgValidatorFCTAddress, ArgValidatorFCTAmount)),
	Run: func(cmd *cobra.Command, args []string) {

	},
}

