// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package opr

import (
	"sync"

	"github.com/pegnet/pegnet/common"
	"github.com/zpatrick/go-config"
)

type IGrader interface {
	GetAlert(id string) (alert chan *OPRs)
	StopAlert(id string)
	Run(config *config.Config, monitor *common.Monitor)
}

// Grader is responsible for evaluating the previous block of OPRs and
// determines who should be paid.
// This also informs the miners which records should be included in their OPR records
type Grader struct {
	alerts      map[string]chan *OPRs
	alertsMutex sync.Mutex // Maps are not thread safe
}

func NewGrader() *Grader {
	g := new(Grader)
	g.alerts = make(map[string]chan *OPRs)

	return g
}

// OPRs is the message sent by the Grader
type OPRs struct {
	ToBePaid []*OraclePriceRecord
	AllOPRs  []*OraclePriceRecord
}

// GetAlert registers a new request for alerts.
// Data will be sent when the grades from the last block are ready
func (g *Grader) GetAlert(id string) (alert chan *OPRs) {
	g.alertsMutex.Lock()
	defer g.alertsMutex.Unlock()

	// If the alert already exists for the id, close it.
	// We only want 1 alert per id
	alert, ok := g.alerts[id]
	if ok {
		close(alert)
	}

	alert = make(chan *OPRs, 10)
	g.alerts[id] = alert
	return g.alerts[id]
}

// StopAlert allows cleanup of alerts that are no longer used
func (g *Grader) StopAlert(id string) {
	g.alertsMutex.Lock()
	defer g.alertsMutex.Unlock()

	alert, ok := g.alerts[id]
	if ok {
		close(alert)
	}
	delete(g.alerts, id)
}

func (g *Grader) Run(config *config.Config, monitor *common.Monitor) {
	InitLX() // We intend to use the LX hash
	fdAlert := monitor.NewListener()
	for {
		fds := <-fdAlert
		if fds.Minute == 1 {
			GetEntryBlocks(config)
			oprs := GetPreviousOPRs(fds.Dbht)
			gradedOPRs, sortedOPRs := GradeBlock(oprs)

			// Alert followers that we have graded the previous block
			g.alertsMutex.Lock() // Lock map to prevent another thread mucking with our loop
			for _, a := range g.alerts {
				var winners OPRs
				winners.ToBePaid = gradedOPRs[:10]
				if len(gradedOPRs) > 10 {
					winners.ToBePaid = gradedOPRs[:10]
				}
				winners.AllOPRs = sortedOPRs
				select { // Don't block if someone isn't pulling from the winner channel
				case a <- &winners:
				default:
					// This means the channel is full
				}
			}
			g.alertsMutex.Unlock()
		}

	}
}
