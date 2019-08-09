package src

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/ratelimit"
)

func init() {
	rootCmd.PersistentFlags().IntP("hashrate", "r", -1, "How many hashes per second to target.")
	rootCmd.PersistentFlags().String("time", "5s", "How long to 'mine' for. This will be for 1 'block'.")
}

// The cli enter point
var rootCmd = &cobra.Command{
	Use:   "simulate",
	Short: "simulate is the cli tool to simulate a higher hashrate",
	Run: func(cmd *cobra.Command, args []string) {
		var _ = cmd.Help()
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		v, _ := cmd.Flags().GetInt("hashrate")
		if v == -1 {
			return
		}
		HashRateLimit = ratelimit.New(v)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
