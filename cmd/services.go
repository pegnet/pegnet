package cmd

import (
	"context"
	"time"

	"github.com/pegnet/pegnet/database"

	"github.com/pegnet/pegnet/api"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/controlPanel"
	"github.com/pegnet/pegnet/mining"
	"github.com/pegnet/pegnet/opr"
	"github.com/zpatrick/go-config"
)

func LaunchFactomMonitor(config *config.Config) *common.Monitor {
	monitor := common.GetMonitor()
	monitor.SetTimeout(time.Duration(Timeout) * time.Second)

	go func() {
		errListener := monitor.NewErrorListener()
		err := <-errListener
		panic("Monitor threw error: " + err.Error())
	}()

	return monitor
}

func LaunchGrader(config *config.Config, monitor *common.Monitor, ctx context.Context) *opr.QuickGrader {
	grader := opr.NewQuickGrader(config, database.NewMapDb())
	go grader.Run(monitor, ctx)
	return grader
}

func LaunchStatistics(config *config.Config, ctx context.Context) *mining.GlobalStatTracker {
	statTracker := mining.NewGlobalStatTracker()

	go statTracker.Collect(ctx) // Will stop collecting on ctx cancel
	return statTracker
}

func LaunchAPI(config *config.Config, stats *mining.GlobalStatTracker, grader *opr.QuickGrader) *api.APIServer {
	s := api.NewApiServer(grader)

	go s.Listen(8099) // TODO: Do not hardcode this
	return s
}

func LaunchControlPanel(config *config.Config, ctx context.Context, monitor common.IMonitor, stats *mining.GlobalStatTracker) *controlPanel.ControlPanel {
	cp := controlPanel.NewControlPanel(config, monitor, stats)
	go cp.ServeControlPanel()
	return cp
}

func LaunchMiners(config *config.Config, ctx context.Context, monitor common.IMonitor, grader opr.IGrader, stats *mining.GlobalStatTracker) *mining.MiningCoordinator {
	coord := mining.NewMiningCoordinatorFromConfig(config, monitor, grader, stats)
	err := coord.InitMinters()
	if err != nil {
		panic(err)
	}

	// TODO: Make this unblocking
	coord.LaunchMiners(ctx) // Inf loop unless context cancelled
	return coord
}
