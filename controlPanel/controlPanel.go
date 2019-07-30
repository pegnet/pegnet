// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package controlPanel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alexandrevicenzi/go-sse"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/mining"
	"github.com/pegnet/pegnet/opr"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

type ControlPanel struct {
	Config     *config.Config
	Statistics *mining.GlobalStatTracker
	Monitor    common.IMonitor

	Server    *http.Server
	SSEServer *sse.Server
}

func corsHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// Our middleware logic goes here...
		next.ServeHTTP(w, r)
	})
}

func NewControlPanel(config *config.Config, monitor common.IMonitor, statTracker *mining.GlobalStatTracker) *ControlPanel {
	c := new(ControlPanel)
	c.Config = config
	c.Monitor = monitor
	c.Statistics = statTracker

	c.Server = &http.Server{}
	// Create the server.
	s := sse.NewServer(&sse.Options{
		// Print debug info
		Logger: nil,
	})

	c.SSEServer = s

	mux := http.NewServeMux()

	// Register with /events endpoint.
	mux.Handle("/events/", c.SSEServer)
	mux.Handle("/", http.FileServer(http.Dir("./controlPanel/static")))
	// GET requests for the CP
	mux.HandleFunc("/cp/miningstats", c.HandleControlPanelRequest)
	c.Server.Handler = corsHeader(mux)

	return c
}

func (c *ControlPanel) Listen(port int) {
	c.Server.Addr = fmt.Sprintf(":%d", port)
	err := c.Server.ListenAndServe()
	if err != nil {
		log.WithError(err).Fatal("control panel stopped")
	}
}

func (c *ControlPanel) Close() {
	var _ = c.Server.Shutdown(context.Background())
	c.SSEServer.Shutdown()
}

func (c *ControlPanel) HandleEvents() {

}

type CommonResponse struct {
	Minute     int64  `json:"minute"`
	Dbht       int32  `json:"dbht"`
	Balance    int64  `json:"balance"`
	HashRate   uint64 `json:"hashRate"`
	Difficulty uint64 `json:"difficulty"`
}

func (c *ControlPanel) ServeControlPanel() {
	log.Info("Starting control panel on localhost:8080")

	alert := c.Monitor.NewListener()
	statsUpStream := c.Statistics.GetUpstream("control-panel")

	network, err := common.LoadConfigNetwork(config)
	if err != nil {
		panic(fmt.Sprintf("Do not have a proper network in the config file: %v", err))
	}

	// Dispatch messages to common channel
	go func() {
		var CurrentHashRate uint64
		var CurrentDifficulty uint64
		var CoinbaseAddress string

		if str, err := c.Config.String(common.ConfigCoinbaseAddress); err != nil {
			log.Fatal("config file has no Coinbase Address")
		} else {
			CoinbaseAddress = str
		}

		CoinbasePNTAddress, err := common.ConvertFCTtoPegNetAsset(network, "PNT", CoinbaseAddress)
		if err != nil {
			panic("no valid coinbase address in the config file")
		}
		// TODO: Include states from statTracker

		for {
			select {
			case e := <-alert:

				hr := common.Stats.GetHashRate()
				diff := common.Stats.Difficulty
				if hr > 0 && hr != CurrentHashRate {
					CurrentHashRate = hr
				}
				if diff > 0 && diff != CurrentDifficulty {
					CurrentDifficulty = diff
				}

				r := CommonResponse{Minute: e.Minute, Dbht: e.Dbht, HashRate: CurrentHashRate, Difficulty: CurrentDifficulty}
				r.Balance = opr.GetBalance(CoinbasePNTAddress)

				data, _ := json.Marshal(r)
				c.SSEServer.SendMessage("/events/common", sse.SimpleMessage(string(data)))
			case s := <-statsUpStream:
				data, _ := json.Marshal(s)
				c.SSEServer.SendMessage("/events/gstats", sse.SimpleMessage(string(data)))
			}
		}
	}()

	c.Listen(8080) // TODO: Do not hardcode

}
