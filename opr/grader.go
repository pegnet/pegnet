// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package opr

import (
	"github.com/pegnet/pegnet/common"
	"github.com/zpatrick/go-config"
)

// Grader is responsible for evaluating the previous block of OPRs and
// determines who should be paid.
// This also informs the miners which records should be included in their OPR records
type Grader struct {
	alerts []chan *OPRs
}

// OPRs is the message sent by the Grader
type OPRs struct {
	ToBePaid []*OraclePriceRecord
	AllOPRs  []*OraclePriceRecord
}

// GetAlert registers a new request for alerts.
// Data will be sent when the grades from the last block are ready
func (g *Grader) GetAlert() chan *OPRs {
	alert := make(chan *OPRs, 10)
	g.alerts = append(g.alerts, alert)
	return alert
}

func (g *Grader) Run(config *config.Config, monitor *common.Monitor) {
	InitLX() // We intend to use the LX hash
	fdAlert := monitor.NewListener()
	for {
		fds := <-fdAlert
		if fds.Minute == 1 {
			GetEntryBlocks(config)
			oprs := GetPreviousOPRs(fds.Dbht)
			tbp, all := GradeBlock(oprs)

			// Alert followers that we have graded the previous block
			for _, a := range g.alerts {
				var winners OPRs
				winners.ToBePaid = tbp
				winners.AllOPRs = all
				a <- &winners
			}
		}

	}
}
