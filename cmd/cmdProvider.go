package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// CmdFlagProvider is able to pull values from the command line, and replace default config
// values.
type CmdFlagProvider struct {
	cmd *cobra.Command
}

func NewCmdFlagProvider(cmd *cobra.Command) *CmdFlagProvider {
	d := new(CmdFlagProvider)
	d.cmd = cmd

	return d
}

func (c *CmdFlagProvider) Load() (map[string]string, error) {
	settings := map[string]string{}

	miners, _ := c.cmd.Flags().GetInt("miners")
	if miners != -1 {
		settings["Miner.NumberOfMiners"] = fmt.Sprintf("%d", miners)
	}

	top, _ := c.cmd.Flags().GetInt("top")
	if top != -1 {
		settings["Miner.RecordsPerBlock"] = fmt.Sprintf("%d", top)
	}

	return settings, nil
}
