package main

import (
	"context"

	"github.com/pegnet/pegnet/balances"
	pcmd "github.com/pegnet/pegnet/cmd"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/node"
	"github.com/spf13/cobra"
)

// This is a bit of a hack atm.
// Because the pegnet node uses sqlite, it adds a dependency for cgo.
// To isolate the dependency to just users that need to use the `node`,
// the `pegnet node` command was brought into this package. Because of the
// 'persistent' use of flags, there are many flags that do not affect the node that will
// be listed. I think for now, this is the simplest change to remove the dependency.

func init() {
	// Reset commands, so `pegnetnode` is the only h
	pcmd.RootCmd.ResetCommands()
	pcmd.RootCmd.Run = func(cmd *cobra.Command, args []string) {
		pegnetNode.Run(cmd, args)
	}
}

// main
func main() {
	pcmd.Execute()
}

var pegnetNode = &cobra.Command{
	Use:   "node",
	Short: "Runs a pegnet node",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		common.GlobalExitHandler.AddCancel(cancel)
		pcmd.ValidateConfig(pcmd.Config) // Will fatal log if it fails
		b := balances.NewBalanceTracker()

		// Services
		monitor := pcmd.LaunchFactomMonitor(pcmd.Config)
		grader := pcmd.LaunchGrader(pcmd.Config, monitor, b, ctx, false)

		pegnetnode, err := node.NewPegnetNode(pcmd.Config, monitor, grader)
		if err != nil {
			pcmd.CmdError(cmd, err)
		}
		common.GlobalExitHandler.AddExit(pegnetnode.Close)

		go pegnetnode.Run(ctx)

		var _ = cancel
		apiserver := pcmd.LaunchAPI(pcmd.Config, nil, grader, b, false)
		apiserver.Mux.Handle("/node/v1", pegnetnode.APIMux())
		// Let's add the pegnet node timeseries to the handle
		apiport, err := pcmd.Config.Int(common.ConfigAPIPort)
		if err != nil {
			pcmd.CmdError(cmd, err)
		}
		apiserver.Listen(apiport)
		var _ = apiserver

	},
}
