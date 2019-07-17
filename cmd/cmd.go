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

	burn.Flags().Bool("dryrun", false, "Dryrun creates the TX without actually submitting it to the network.")
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
	Use:   "burn <fct address> <fct amount>",
	Short: "Burns the specied amount of FCT into PNT",
	Long: "Burning FCT will turn it into PNT. The PNT burn address is an EC address, and the transaction has " +
		"an input with # of FCT, and an output of 0 EC. This means the entire tx input becomes the fee. " +
		"This command costs FCT, so be careful when using it.",
	Example: "pegnet burn FA3EPZYqodgyEGXNMbiZKE5TS2x2J9wF8J9MvPZb52iGR78xMgCb 1",
	// TODO: Verify this functionality.
	ValidArgs: ValidOwnedFCTAddresses(),
	Args:      CombineCobraArgs(CustomArgOrderValidationBuilder(true, ArgValidatorFCTAddress, ArgValidatorFCTAmount)),
	Run: func(cmd *cobra.Command, args []string) {
		defer factom.DeleteTransaction("burn") // Any cleanup from errors
		// First see if we own the specified FCT address
		_, err := factom.FetchFactoidAddress(args[0])
		if err != nil {
			fmt.Println(err)
			return
		}

		// Get our balance
		factoshiBalance, err := factom.GetFactoidBalance(args[0])
		if err != nil {
			fmt.Println(err)
			return
		}

		// Ensure our balance is enough to cover the burn
		factoshiBurn := factom.FactoidToFactoshi(args[1])
		if factoshiBurn > uint64(factoshiBalance) {
			fctBal := factom.FactoshiToFactoid(uint64(factoshiBalance))
			fmt.Printf("You only have %s FCT, you specified to burn %s\n", fctBal, args[1])
			return
		}

		name := "burn"
		if _, err := factom.NewTransaction(name); err != nil {
			fmt.Println(err)
			return
		}
		if _, err := factom.AddTransactionInput(name, args[0], factoshiBurn); err != nil {
			fmt.Println(err)
			return
		}
		if _, err := factom.AddTransactionECOutput(name, common.PegnetBurnAddress(Network), 0); err != nil {
			fmt.Println(err)
			return
		}

		// Signing the tx without a force makes the wallet check the fee amount
		if _, err := factom.SignTransaction(name, false); err != nil {
			// Only care about the insufficient fee error here
			if strings.Contains(err.Error(), "Insufficient Fee") {
				fmt.Println(err)
				return
			}
		}

		// We will force the transaction to ignore any fee too high errors
		if _, err := factom.SignTransaction(name, true); err != nil {
			fmt.Println(err)
			return
		}

		if dryrun, _ := cmd.Flags().GetBool("dryrun"); dryrun {
			tx, err := factom.ComposeTransaction(name)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("This transaction was not submitted to the network.")
			fmt.Println(string(tx))
			return
		}
		tx, err := factom.SendTransaction(name)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Burn transaction sent to the network")
		fmt.Printf("TransacitonID: %s\n", tx.TxID)
	},
}
