package opr

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/pegnet/pegnet/common"

	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

var recLog = log.WithFields(log.Fields{"id": "recorder"})

// ChainRecorder is a tool to create csvs to look at things on chain
type ChainRecorder struct {
	config   *config.Config
	filepath string
}

func NewChainRecorder(con *config.Config, filpath string) (*ChainRecorder, error) {
	c := new(ChainRecorder)
	c.config = con
	c.filepath = filpath

	return c, nil
}

func (c *ChainRecorder) WriteMinerCSV() error {
	InitLX() // We intend to use the LX hash

	f, err := os.OpenFile(c.filepath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	file := f
	writer := csv.NewWriter(file)

	recLog.Infof("fetching entry blocks")
	for tries := 0; tries < 3; tries++ {
		err = GetEntryBlocks(c.config)
		if err == nil {
			break
		} else {
			// If this fails, we probably can't recover this block.
			// Can't hurt to try though
			time.Sleep(200 * time.Millisecond)
		}
	}

	if err != nil {
		return err
	}

	recLog.WithField("blockcount", len(OPRBlocks)).Infof("writing to csv")
	cutoff, _ := c.config.Int(common.ConfigSubmissionCutOff)
	err = writer.Write([]string{
		"blockheight", "records",
		"1st Diff (int)", "1st Diff (hex)",
		"Last Graded Place", "Last Graded Diff (int)", "Last Graded Diff (hex)",
		"Cutoff", fmt.Sprintf("%d cutoff (int)", cutoff), fmt.Sprintf("%d cutoff (hex)", cutoff),
		"Last OPR Place", "Last OPR (int)", "Last OPR (hex)",
	}) // Write headers
	if err != nil {
		return err
	}

	// Build the csv
	for i, block := range OPRBlocks {
		var _ = i
		last := 50
		if len(block.OPRs) < 50 {
			last = len(block.OPRs) - 1
		}
		if last < 0 {
			continue
		}

		cutoffD := CalculateMinimumDifficultyFromOPRs(block.OPRs, cutoff)

		err = writer.Write([]string{
			fmt.Sprintf("%d", block.Dbht),
			fmt.Sprintf("%d", len(block.OPRs)),

			fmt.Sprintf("%d", block.OPRs[0].Difficulty),
			fmt.Sprintf("%x", block.OPRs[0].Difficulty),

			fmt.Sprintf("%d", last),
			fmt.Sprintf("%d", block.OPRs[last].Difficulty),
			fmt.Sprintf("%x", block.OPRs[last].Difficulty),

			fmt.Sprintf("%d", cutoff),
			fmt.Sprintf("%d", cutoffD),
			fmt.Sprintf("%x", cutoffD),

			fmt.Sprintf("%d", len(block.OPRs)),
			fmt.Sprintf("%d", block.OPRs[len(block.OPRs)-1].Difficulty),
			fmt.Sprintf("%x", block.OPRs[len(block.OPRs)-1].Difficulty),
		})
		if err != nil {
			return err
		}
	}

	writer.Flush()
	var _ = file.Close()

	return nil
}
