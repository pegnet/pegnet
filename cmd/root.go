// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"regexp"
	"strings"
	"time"

	"github.com/pegnet/pegnet/mining"

	"github.com/FactomProject/factom"
	"github.com/pegnet/pegnet/api"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/controlPanel"
	"github.com/pegnet/pegnet/opr"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zpatrick/go-config"
)

var (
	Config *config.Config
	// Global Flags
	LogLevel        string
	FactomdLocation string
	WalletdLocation string
	Timeout         uint
	Network         string
)

func init() {
	// TODO: Review how this completion works
	//		The autotab stuff doesn't update automatically
	rootCmd.AddCommand(completionCmd)

	rootCmd.PersistentFlags().StringVar(&LogLevel, "log", "info", "Change the logging level. Can choose from 'trace', 'debug', 'info', 'warn', 'error', or 'fatal'")
	rootCmd.PersistentFlags().StringVar(&FactomdLocation, "s", "localhost:8088", "IPAddr:port# of factomd API to use to access blockchain (default localhost:8088)")
	rootCmd.PersistentFlags().StringVar(&WalletdLocation, "w", "localhost:8089", "IPAddr:port# of factom-walletd API to use to create transactions (default localhost:8089)")
	rootCmd.PersistentFlags().UintVar(&Timeout, "timeout", 90, "The time (in seconds) that the miner tolerates the downtime of the factomd API before shutting down")
	rootCmd.PersistentFlags().StringVar(&Network, "network", "test", "The pegnet network to target. <Main|Test>")

	// Flags that affect the config file. Should not be loaded into globals
	rootCmd.PersistentFlags().Int("miners", -1, "Change the number of miners being run (default to config file)")
	rootCmd.PersistentFlags().Int("top", -1, "Change the number opr records written per block (default to config file)")

	// Persist flags that run in PreRun, and not in the config
	rootCmd.PersistentFlags().Bool("profile", false, "GoLang profiling")
	rootCmd.PersistentFlags().Int("profileport", 7060, "Change profiling port (default 16060)")

	// Initialize the config file with the config, then with cmd flags
	rootCmd.PersistentPreRunE = rootPreRunSetup

	// Run a few functions (in the order specified) to initialize some globals
	cobra.OnInitialize(initLogger, initFactomdLocs)
}

// The cli enter point
var rootCmd = &cobra.Command{
	Use:   "pegnet",
	Short: "pegnet is the cli tool to run or interact with a PegNet node",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Do we really want to init the miner by default?
		//  	 Not like `pegnet -service=miner` or something?

		ValidateConfig(Config) // Will fatal log if it fails

		// Default to config options if cli flags aren't specified
		//miners, _ := Config.Int("Miner.NumberOfMiners")

		monitor := common.GetMonitor()
		monitor.SetTimeout(time.Duration(Timeout) * time.Second)

		go func() {
			errListener := monitor.NewErrorListener()
			err := <-errListener
			panic("Monitor threw error: " + err.Error())
		}()

		grader := opr.NewGrader()
		go grader.Run(Config, monitor)

		http.Handle("/v1", api.RequestHandler{})
		go http.ListenAndServe(":8099", nil)

		statTracker := mining.NewGlobalStatTracker()
		go controlPanel.ServeControlPanel(Config, monitor, statTracker)

		coord := mining.NewMiningCoordinatorFromConfig(Config, monitor, grader, statTracker)
		err := coord.InitMinters()
		if err != nil {
			panic(err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		go statTracker.Collect(ctx) // Will stop collecting on ctx cancel
		coord.LaunchMiners(context.Background())

		// Calling cancel() will cancel the stat tracker collection AND the miners
		var _ = cancel

		//if miners > 0 {
		//	miners := pegnetMining.InitMiners(Config, monitor, grader)
		//	for _, m := range miners[:len(miners)-1] {
		//		go m.LaunchMiningThread(false)
		//	}
		//	miners[len(miners)-1].LaunchMiningThread(true) // Launch last one to hog thread
		//}
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

	valid, _ := regexp.MatchString("^[a-zA-Z0-9,]+$", identity)
	if !valid {
		log.WithError(err).Fatal("only alphanumeric characters and commas are allowed in the identity")
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

func initFactomdLocs() {
	factom.SetFactomdServer(FactomdLocation)
	factom.SetWalletServer(WalletdLocation)
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
	userPath := u.HomeDir
	configFile := fmt.Sprintf("%s/.%s/defaultconfig.ini", userPath, "pegnet")
	iniFile := config.NewINIFile(configFile)
	flags := NewCmdFlagProvider(cmd)
	Config = config.NewConfig([]config.Provider{iniFile, flags})

	// Profiling setup
	if on, _ := cmd.Flags().GetBool("profile"); on {
		p, _ := cmd.Flags().GetInt("profileport")
		go StartProfiler(p)
	}

	return nil
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

func CmdError(cmd *cobra.Command, i interface{}) {
	cmd.PrintErr(i)
	os.Exit(1)
}

func CmdErrorf(cmd *cobra.Command, format string, i ...interface{}) {
	cmd.PrintErrf(format, i...)
	os.Exit(1)
}
