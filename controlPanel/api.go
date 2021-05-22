package controlPanel

import (
	"net/http"

	"github.com/pegnet/pegnet/api"
)

/*
 * Control panel specific API endpoints
 * These can change without having to spec out the api as much
 */

type StatisticAPIRequest struct {
	BlockRange api.BlockRange `json:"block_range"`
}

func (g *ControlPanel) HandleControlPanelRequest(w http.ResponseWriter, r *http.Request) {
	// TODO: parse the range and actually fufill the request from the GET uri

	// Currenty just default to all
	s := g.Statistics
	stats := s.FetchAllStats()

	api.Respond(w, api.PostResponse{Res: stats})
}
