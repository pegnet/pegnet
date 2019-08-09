package src

import (
	"fmt"
	"math/rand"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/ratelimit"
)

func init() {
	rootCmd.PersistentFlags().IntP("hashrate", "r", -1, "How many hashes per second to target.")
	rootCmd.PersistentFlags().String("time", "5s", "How long to 'mine' for. This will be for 1 'block'.")
	rootCmd.PersistentFlags().Float64("maxflux", 0, "Fluctuates hashrate from the base by maximum  this % each block")
	rootCmd.PersistentFlags().Float64("minflux", 0, "Fluctuates hashrate from the base by minimum this % each block")
}

// The cli enter point
var rootCmd = &cobra.Command{
	Use:   "simulate",
	Short: "simulate is the cli tool to simulate a higher hashrate",
	Run: func(cmd *cobra.Command, args []string) {
		var _ = cmd.Help()
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		UpdateHashRate(cmd, args)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func UpdateHashRate(cmd *cobra.Command, args []string) {
	v, _ := cmd.Flags().GetInt("hashrate")
	if v == -1 {
		return
	}

	// Check for flux
	max, _ := cmd.Flags().GetFloat64("maxflux")
	min, _ := cmd.Flags().GetFloat64("minflux")
	if max != 0 || min != 0 {
		r := min + rand.Float64()*(max-min)
		// As a percent + 1
		r = 1 + (r / 100)

		v = int(float64(v) * r)
	} else {
		// Do not set a flux, if we already set this, leave it
		if HashRateLimit != nil {
			return
		}
	}

	HashRateLimit = ratelimit.New(v)
}
