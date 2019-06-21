package opr

import (
	"github.com/pegnet/pegnet/support"
	"github.com/zpatrick/go-config"
)

// We have one grader that evaluates the previous block of OPRs and determines who should be paid
// This also informs the miners what records should be included in their OPR records
type Grader struct {
	alerts []chan *OPRs
}

// Alert from the grading service
type OPRs struct {
	ToBePaid []*OraclePriceRecord
	AllOPRs  []*OraclePriceRecord
}

// Miners sign up to be alerted when the grades from the last block are ready
func (g *Grader) GetAlert() chan *OPRs {
	alert := make(chan *OPRs, 10)
	g.alerts = append(g.alerts, alert)
	return alert
}

func (g *Grader) Run(config *config.Config, monitor *support.FactomdMonitor) {
	fdAlert := monitor.GetAlert()
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
