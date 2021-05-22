package cmd

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/pegnet/pegnet/common"

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
	var err error
	settings := map[string]string{}

	miners, _ := c.cmd.Flags().GetInt("miners")
	if miners != -1 {
		settings["Miner.NumberOfMiners"] = fmt.Sprintf("%d", miners)
	}

	top, _ := c.cmd.Flags().GetInt("top")
	if top != -1 {
		settings["Miner.RecordsPerBlock"] = fmt.Sprintf("%d", top)
	}

	id, _ := c.cmd.Flags().GetString("identity")
	if id != "" {
		settings["Miner.IdentityChain"] = id
	}

	factomd, _ := c.cmd.Flags().GetString("factomdlocation")
	if factomd != "localhost:8088" {
		settings["Miner.FactomdLocation"] = factomd
	}
	walletd, _ := c.cmd.Flags().GetString("walletdlocation")
	if walletd != "localhost:8089" {
		settings["Miner.WalletdLocation"] = walletd
	}

	// Use the same flag for the client and server.
	addr, _ := c.cmd.Flags().GetString("caddr")
	if addr != "" {
		settings[common.ConfigCoordinatorLocation] = addr
		settings[common.ConfigCoordinatorListen] = addr
	}

	network, _ := c.cmd.Flags().GetString("network")
	if network != "" {
		settings[common.ConfigPegnetNetwork], err = common.GetNetwork(network)
		if err != nil {
			return settings, err
		}
	}

	dbtype, _ := c.cmd.Flags().GetString("minerdbtype")
	if dbtype != "" {
		settings[common.ConfigMinerDBType] = dbtype
	}

	dbpath, _ := c.cmd.Flags().GetString("minerdb")
	if dbpath != "" {
		settings[common.ConfigMinerDBPath] = dbpath
	}

	// Overrride
	arr, _ := c.cmd.Flags().GetStringArray("override")
	for _, v := range arr {
		values := strings.Split(v, "=")
		if len(values) == 2 {
			settings[values[0]] = values[1]
		} else {
			logrus.Warn("Override fields expect format as 'Section.Option=Value'")
		}
	}

	return settings, nil
}
