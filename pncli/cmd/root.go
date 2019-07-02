package cmd

import (
	"fmt"
	"os"

	"strings"

	"github.com/FactomProject/factom"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// Global Flags
	LogLevel        string // Logrus global log level
	FactomdLocation string
	WalletdLocation string
)

func init() {
	// TODO: Review how this completion works
	//		The autotab stuff doesn't update automatically
	rootCmd.AddCommand(completionCmd)

	rootCmd.PersistentFlags().StringVar(&LogLevel, "log", "info", "Change the logging level. Can choose from 'debug', 'info', 'warn', 'error', or 'fatal'")
	rootCmd.PersistentFlags().StringVar(&FactomdLocation, "s", "localhost:8088", "IPAddr:port# of factomd API to use to access blockchain (default localhost:8088)")
	rootCmd.PersistentFlags().StringVar(&WalletdLocation, "w", "localhost:8089", "IPAddr:port# of factom-walletd API to use to create transactions (default localhost:8089)")

	// Always init the logrus global logger
	cobra.OnInitialize(initLogger, initFactomdLocs)
}

// The cli enter point
var rootCmd = &cobra.Command{
	Use:   "pncli",
	Short: "pncli is the cli tool to interact with the pegnet daemon",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {

		fmt.Println(err)
		os.Exit(1)
	}
}

func initFactomdLocs() {
	factom.SetFactomdServer(FactomdLocation)
	factom.SetWalletServer(WalletdLocation)
}

func initLogger() {
	switch strings.ToLower(LogLevel) {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	}
}

// https://github.com/spf13/cobra/blob/master/bash_completions.md
// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "!EXPERIMENTAL! Generates bash completion scripts.",
	Long: `!EXPERIMENTAL! Generates bash completion scripts. You can store something like this in your bashrc: 
pncli completion > /tmp/ntc && source /tmp/ntc`,
	Run: func(cmd *cobra.Command, args []string) {
		addGetEncodingCommands()
		rootCmd.GenBashCompletion(os.Stdout)
	},
}
