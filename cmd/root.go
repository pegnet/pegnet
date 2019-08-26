// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/FactomProject/factom"
	"github.com/pegnet/pegnet/balances"
	"github.com/pegnet/pegnet/common"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zpatrick/go-config"
)

var (
	Config      *config.Config
	ExitHandler *common.ExitHandler
	// Global Flags
	LogLevel        string
	FactomdLocation string
	WalletdLocation string
	Timeout         uint
)

func init() {
	// TODO: Review how this completion works
	//		The autotab stuff doesn't update automatically
	rootCmd.AddCommand(completionCmd)

	rootCmd.PersistentFlags().StringVar(&LogLevel, "log", "info", "Change the logging level. Can choose from 'trace', 'debug', 'info', 'warn', 'error', or 'fatal'")
	rootCmd.PersistentFlags().StringVarP(&FactomdLocation, "factomdlocation", "s", "localhost:8088", "IPAddr:port# of factomd API to use to access blockchain")
	rootCmd.PersistentFlags().StringVarP(&WalletdLocation, "walletdlocation", "w", "localhost:8089", "IPAddr:port# of factom-walletd API to use to create transactions")
	rootCmd.PersistentFlags().UintVar(&Timeout, "timeout", 90, "The time (in seconds) that the miner tolerates the downtime of the factomd API before shutting down")

	// Flags that affect the config file. Should not be loaded into globals
	rootCmd.PersistentFlags().Int("miners", -1, "Change the number of miners being run (default to config file)")
	rootCmd.PersistentFlags().Int("top", -1, "Change the number opr records written per block (default to config file)")
	rootCmd.PersistentFlags().String("identity", "", "Change the identity being used (default to config file)")
	rootCmd.PersistentFlags().String("caddr", "", "Change the location of the coordinator. (default to config file)")
	rootCmd.PersistentFlags().String("config", "", "Set a custom filepath for the config file. (default is ~/.pegnet/defaultconfig.ini)")
	rootCmd.PersistentFlags().String("mienrdb", "", "Set a custom filepath for the miner database. (default is ~/.pegnet/miner.ldb)")
	rootCmd.PersistentFlags().String("mienrdbtype", "", "Set the db type for the miner. (default is ~/.pegnet/miner.ldb)")

	// Persist flags that run in PreRun, and not in the config
	rootCmd.PersistentFlags().Bool("profile", false, "GoLang profiling")
	rootCmd.PersistentFlags().Int("profileport", 7060, "Change profiling port (default 16060)")
	rootCmd.PersistentFlags().String("network", "", "The pegnet network to target. <MainNet|TestNet>")

	// Initialize the config file with the config, then with cmd flags
	rootCmd.PersistentPreRunE = rootPreRunSetup

	// Run a few functions (in the order specified) to initialize some globals
	cobra.OnInitialize(initLogger)
}

// The cli enter point
var rootCmd = &cobra.Command{
	Use:   "pegnet",
	Short: "pegnet is the cli tool to run or interact with a PegNet node",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		b := balances.NewBalanceTracker()

		ValidateConfig(Config) // Will fatal log if it fails

		// Services
		monitor := LaunchFactomMonitor(Config)
		grader := LaunchGrader(Config, monitor, b, ctx, true)
		statTracker := LaunchStatistics(Config, ctx)
		apiserver := LaunchAPI(Config, statTracker, grader, b, true)
		LaunchControlPanel(Config, ctx, monitor, statTracker, b)
		var _ = apiserver

		// This is a blocking call
		coord := LaunchMiners(Config, ctx, monitor, grader, statTracker)

		// Calling cancel() will cancel the stat tracker collection AND the miners
		var _, _ = cancel, coord
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// ValidateConfig will validate the config is up to snuff.
// Do w/e config validation we want. Will fatal if it fails
func ValidateConfig(config *config.Config) {
	_, err := config.String("Miner.Protocol")
	if err != nil {
		log.WithError(err).Fatal("failed to read miner protocol from config")
	}
	_, err = config.Int("Miner.NumberOfMiners")
	if err != nil {
		log.WithError(err).Fatal("failed to read number of miners from config")
	}

	identity, err := config.String("Miner.IdentityChain")
	if err != nil {
		log.WithError(err).Fatal("failed to read the identity chain or miner id")
	}

	if err := common.ValidIdentity(identity); err != nil {
		log.WithError(err).Fatal("invalid identity")
	}

}

func initLogger() {
	switch strings.ToLower(LogLevel) {
	case "trace":
		log.SetLevel(log.TraceLevel)
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

// rootPreRunSetup is run before all cmd commands. It will:
//		1: Parse the config
//		2: Parse the cmd flags that overwrite the config
//		3. Launch profiling if we have it enabled
func rootPreRunSetup(cmd *cobra.Command, args []string) error {
	// Config setup
	u, err := user.Current()
	if err != nil {
		log.WithError(err).Fatal("Failed to read current user's name")
	}

	// Set the PegnetHome for pegnet files.
	// This is so we know where to place all the pegnet files. We can then use
	// the $PEGNETHOME in place of ~/.pegnet and be able to change it in only 1 spot.
	pegnethome := os.Getenv("PEGNETHOME")
	if pegnethome == "" {
		var _ = os.Setenv("PEGNETHOME", filepath.Join(u.HomeDir, ".pegnet"))
	}

	configFile := os.ExpandEnv(filepath.Join("$PEGNETHOME", "defaultconfig.ini"))
	customPath, _ := cmd.Flags().GetString("config")
	if customPath != "" {
		absPath, err := filepath.Abs(customPath)
		if err == nil {
			configFile = absPath
			log.Info("Using config file: ", configFile)
		}
	}

	iniFile := config.NewINIFile(configFile)
	flags := NewCmdFlagProvider(cmd)
	Config = config.NewConfig([]config.Provider{common.NewDefaultConfigOptionsProvider(), iniFile, flags})

	pegnetnetwork := os.Getenv("PEGNETNETWORK")
	if pegnetnetwork == "" {
		net, err := common.LoadConfigNetwork(Config)
		if err != nil {
			return err
		}
		var _ = os.Setenv("PEGNETNETWORK", net)
	}

	factomd, _ := Config.String("Miner.FactomdLocation")
	walletd, _ := Config.String("Miner.WalletdLocation")
	factom.SetFactomdServer(factomd)
	factom.SetWalletServer(walletd)

	// Profiling setup
	if on, _ := cmd.Flags().GetBool("profile"); on {
		p, _ := cmd.Flags().GetInt("profileport")
		go StartProfiler(p)
	}

	// Catch ctl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		log.Info("Gracefully closing")
		common.GlobalExitHandler.Close()

		log.Info("closing application")
		// If something is hanging, we have to kill it
		os.Exit(0)
	}()

	return nil
}

// https://github.com/spf13/cobra/blob/master/bash_completions.md
// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "!EXPERIMENTAL! Generates bash completion scripts.",
	Long: `!EXPERIMENTAL! Generates bash completion scripts. You can store something like this in your bashrc: 
pegnet completion > /tmp/ntc && source /tmp/ntc`,
	Run: func(cmd *cobra.Command, args []string) {
		addGetEncodingCommands()
		rootCmd.GenBashCompletion(os.Stdout)
	},
}

func CmdError(cmd *cobra.Command, i interface{}) {
	cmd.PrintErr(i)
	os.Exit(1)
}

func CmdErrorf(cmd *cobra.Command, format string, i ...interface{}) {
	cmd.PrintErrf(format, i...)
	os.Exit(1)
}
