package cmd

import (
	"context"
	"github.com/pegnet/pegnet/staking"
	"os"
	"strings"
	"time"

	"github.com/pegnet/pegnet/balances"

	"github.com/pegnet/pegnet/api"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/controlPanel"
	"github.com/pegnet/pegnet/database"
	"github.com/pegnet/pegnet/mining"
	"github.com/pegnet/pegnet/opr"
	log "github.com/sirupsen/logrus"
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

func LaunchGrader(config *config.Config, monitor *common.Monitor, balances *balances.BalanceTracker, ctx context.Context, run bool) *opr.QuickGrader {
	db := OpenDB(Config)

	grader := opr.NewQuickGrader(config, db, balances)
	if run {
		go grader.Run(monitor, ctx)
	}
	common.GlobalExitHandler.AddExit(grader.Close)
	return grader
}

func OpenDB(config *config.Config) database.IDatabase {
	dbtype, err := config.String(common.ConfigMinerDBType)
	if err != nil {
		log.WithError(err).Fatal("Database.MinerDatabaseType needs to be set in the config file or cmd line")
		os.Exit(1)
	}

	var db database.IDatabase

	switch strings.ToLower(dbtype) {
	case "map":
		db = database.NewMapDb()
	case "ldb":
		db = OpenLevelDB(config)
	default:
		log.Fatalf("%s is not a valid db type", dbtype)
		os.Exit(1)
	}
	return db
}

func OpenLevelDB(config *config.Config) *database.Ldb {
	dbpath, err := config.String(common.ConfigMinerDBPath)
	if err != nil {
		log.WithError(err).Fatal("Database.MinerDatabase needs to be set in the config file or cmd line")
		os.Exit(1)
	}

	ldb := new(database.Ldb)
	err = ldb.Open(os.ExpandEnv(dbpath))
	if err != nil {
		log.WithError(err).Fatal("ldb failed to open")
		os.Exit(1)
	}
	return ldb
}

func LaunchStatistics(config *config.Config, ctx context.Context) *mining.GlobalStatTracker {
	statTracker := mining.NewGlobalStatTracker()

	go statTracker.Collect(ctx) // Will stop collecting on ctx cancel
	return statTracker
}

func LaunchAPI(config *config.Config, stats *mining.GlobalStatTracker, grader *opr.QuickGrader, bals *balances.BalanceTracker, run bool) *api.APIServer {
	s := api.NewApiServer(grader, bals, config)

	if run {
		apiport, err := config.Int(common.ConfigAPIPort)
		if err != nil {
			log.WithError(err).Fatal("can't find api port")
			os.Exit(1)
		}
		go s.Listen(apiport)
	}
	return s
}

func LaunchControlPanel(config *config.Config, ctx context.Context, monitor common.IMonitor, stats *mining.GlobalStatTracker, bals *balances.BalanceTracker) *controlPanel.ControlPanel {
	cp := controlPanel.NewControlPanel(config, monitor, stats, bals)
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

func LaunchStaker(config *config.Config, ctx context.Context, monitor common.IMonitor) *staking.StakingCoordinator {
	coord_s := staking.NewStakingCoordinatorFromConfig(config, monitor)
	err := coord_s.InitStaker()
	if err != nil {
		panic(err)
	}

	coord_s.LaunchStaker(ctx) // Inf loop unless context cancelled
	return coord_s
}