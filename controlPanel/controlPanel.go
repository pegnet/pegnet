package controlPanel

import (
	"encoding/json"
	"net/http"

	"github.com/alexandrevicenzi/go-sse"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/mining"
	"github.com/pegnet/pegnet/opr"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

type CommonResponse struct {
	Minute     int64  `json:"minute"`
	Dbht       int32  `json:"dbht"`
	Balance    int64  `json:"balance"`
	HashRate   uint64 `json:"hashRate"`
	Difficulty uint64 `json:"difficulty"`
}

func ServeControlPanel(config *config.Config, monitor *common.Monitor, statTracker *mining.GlobalStatTracker) {
	log.Info("Starting control panel on localhost:8080")

	alert := monitor.NewListener()

	// Create the server.
	s := sse.NewServer(&sse.Options{
		// Print debug info
		Logger: nil,
	})
	defer s.Shutdown()

	// Register with /events endpoint.
	http.Handle("/events/", s)

	// Dispatch messages to common channel
	go func() {
		var CurrentHashRate uint64
		var CurrentDifficulty uint64
		var CoinbasePNTAddress string

		if str, err := config.String("Miner.CoinbasePNTAddress"); err != nil {
			log.Fatal("config file has no Coinbase PNT Address")
		} else {
			CoinbasePNTAddress = str
		}

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
				s.SendMessage("/events/common", sse.SimpleMessage(string(data)))
			}
		}
	}()

	http.Handle("/", http.FileServer(http.Dir("./controlPanel/static")))
	http.ListenAndServe(":8080", nil)

}
