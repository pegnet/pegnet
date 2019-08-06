// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package opr

import (
	"fmt"
	"sync"
	"time"

	"github.com/pegnet/pegnet/common"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

var gLog = log.WithField("id", "grader")

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

	// Since this is used as a message, we need a way to send an error
	Error error
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
		fLog := gLog.WithFields(log.Fields{"minute": fds.Minute, "dbht": fds.Dbht})
		if fds.Minute == 1 {
			var err error
			tries := 0
			// Try 3 times
			for tries = 0; tries < 3; tries++ {
				err = nil
				err = GetEntryBlocks(config)
				if err == nil {
					break
				}
				if err != nil {
					// If this fails, we probably can't recover this block.
					// Can't hurt to try though
					time.Sleep(200 * time.Millisecond)
				}
			}

			if err != nil {
				fLog.WithError(err).WithField("tries", tries).Errorf("Grader failed to grade blocks. Sitting out this block")
				g.SendToListeners(&OPRs{Error: fmt.Errorf("failed to grade")})
				continue
			}

			oprs := GetPreviousOPRs(fds.Dbht)
			gradedOPRs, sortedOPRs := GradeBlock(oprs)

			var winners OPRs
			if len(gradedOPRs) >= 10 {
				winners.ToBePaid = gradedOPRs[:10]
			}
			winners.AllOPRs = sortedOPRs

			// Alert followers that we have graded the previous block
			g.SendToListeners(&winners)
		}

	}
}

func (g *Grader) SendToListeners(winners *OPRs) {
	g.alertsMutex.Lock() // Lock map to prevent another thread mucking with our loop
	for _, a := range g.alerts {
		select { // Don't block if someone isn't pulling from the winner channel
		case a <- winners:
		default:
			// This means the channel is full
		}
	}
	g.alertsMutex.Unlock()
}
