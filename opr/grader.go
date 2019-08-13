// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package opr

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/FactomProject/factom"

	"github.com/pegnet/pegnet/common"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

var gLog = log.WithField("id", "grader")

type IGrader interface {
	GetAlert(id string) (alert chan *OPRs)
	StopAlert(id string)
	Run(monitor *common.Monitor, ctx context.Context)
}

// QuickGrader is responsible for evaluating the previous block of OPRs and
// determines who should be paid. It will only check the minimum number of
// OPR records.
// This also informs the miners which records should be included in their OPR records
type QuickGrader struct {
	Network          string
	Protocol         string
	OPRChainID       []byte
	OPRChainIDString string

	OPRChain *EntryBlockSync

	Config *config.Config

	// oprblks is all the eblocks that contain the oprs
	oprblks []*OprBlock
	// lastGraded is the last graded oprblk, so we know
	// where to start grading
	lastGraded int

	alerts      map[string]chan *OPRs
	alertsMutex sync.Mutex // Maps are not thread safe
}

func NewQuickGrader(config *config.Config) *QuickGrader {
	InitLX()
	g := new(QuickGrader)
	g.Config = config

	network, err := common.LoadConfigNetwork(config)
	common.CheckAndPanic(err)
	p, err := config.String("Miner.Protocol")
	common.CheckAndPanic(err)

	opr := [][]byte{[]byte(p), []byte(network), []byte(common.OPRChainTag)}

	g.Network = network
	g.Protocol = p
	g.OPRChainID = common.ComputeChainIDFromFields(opr)
	g.OPRChainIDString = hex.EncodeToString(g.OPRChainID)

	g.alerts = make(map[string]chan *OPRs)

	g.OPRChain = NewEntryBlockSync(g.OPRChainIDString)
	g.oprblks = make([]*OprBlock, 0)

	return g
}

// GetBlocks should only be used in unit tests. It is not thread safe
func (g *QuickGrader) GetBlocks() []*OprBlock {
	return g.oprblks
}

// GetAlert registers a new request for alerts.
// Data will be sent when the grades from the last block are ready
func (g *QuickGrader) GetAlert(id string) (alert chan *OPRs) {
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
func (g *QuickGrader) StopAlert(id string) {
	g.alertsMutex.Lock()
	defer g.alertsMutex.Unlock()

	alert, ok := g.alerts[id]
	if ok {
		close(alert)
	}
	delete(g.alerts, id)
}

func (g *QuickGrader) Run(monitor *common.Monitor, ctx context.Context) {
	InitLX() // We intend to use the LX hash
	for {    // We need to first sync our grader before we start syncing new blocks
		select { // If we get stuck in the sync loop, this is how can cancel it
		case <-ctx.Done():
			return // Grader stopped
		default:
		}
		err := g.Sync() // Might want to pass a context down?
		if err != nil { // We will try again in a little bit
			log.WithField("id", "grader").WithError(err).Errorf("failed to sync")
			time.Sleep(2 * time.Second)
			continue
		}
		break // Initial sync done
	}

	fdAlert := monitor.NewListener()
	for {
		var fds common.MonitorEvent
		select {
		case fds = <-fdAlert:
		case <-ctx.Done():
			return // Grader stopped
		}
		fLog := gLog.WithFields(log.Fields{"minute": fds.Minute, "dbht": fds.Dbht})
		if fds.Minute == 1 {
			var err error
			tries := 0
			// Try 3 times
			for tries = 0; tries < 3; tries++ {
				err = nil
				err = g.Sync()
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

			oprs := g.GetPreviousOPRBlock(fds.Dbht)

			var winners OPRs
			winners.ToBePaid = oprs.GradedOPRs[:10]
			winners.AllOPRs = oprs.OPRs

			// Alert followers that we have graded the previous block
			g.SendToListeners(&winners)
		}

	}
}

// Sync will sync our opr chain to the latest eblock head of the OPR chain
func (g *QuickGrader) Sync() error {
	// Syncblocks will take our chain and gather all the eblocks
	// that might remain to be synced. This means this function ONLY syncs eblocks
	// and from there we can sync the blocks one by one
	err := g.OPRChain.SyncBlocks()
	if err != nil {
		return err
	}

	// If we have eblocks to sync, this is where we go through them one by one.
	if !g.OPRChain.Synced() {
		// We have eblocks to sync!
		// NextEBlock() will return the next eblock in the chain that we still need to sync
		// it's entries. So we can walk, NextEblock -> NextEblock -> ... to sync the whole chain.
		for block := g.OPRChain.NextEBlock(); block != nil; block = g.OPRChain.NextEBlock() {
			// There is not enough entries in this block, so there is no point in looking at it.
			if len(block.EntryBlock.EntryList) < 10 {
				g.OPRChain.BlockParsed(*block) // This block is done being processed
				continue
			}

			var oprs []*OraclePriceRecord
			var err error
			if len(block.EntryBlock.EntryList) > 50 {
				// Multithread when there is a lot (like a 6x speedup in my tests against mainnet)
				oprs, err = g.ParallelFetchOPRsFromEBlock(block)
			} else {
				oprs, err = g.FetchOPRsFromEBlock(block)
			}
			if err != nil {
				return err
			}

			// Check if we have enough oprs for this block. If we have less than 10, there is no point in
			// trying to grade it.
			if len(oprs) < 10 {
				g.OPRChain.BlockParsed(*block) // This block is done being processed
				continue                       // Not enough oprs for this block be a valid oprblock
			}

			// Sort the OPRs by self reported difficulty
			// We will toss dishonest ones when we grade.
			sort.SliceStable(oprs, func(i, j int) bool {
				return binary.BigEndian.Uint64(oprs[i].SelfReportedDifficulty) > binary.BigEndian.Uint64(oprs[j].SelfReportedDifficulty)
			})

			// GradeMinimum will only grade the first 50 honest records
			graded := GradeMinimum(oprs)
			if len(graded) < 10 { // We might lose some when we reject dishonest records
				g.OPRChain.BlockParsed(*block) // This block is done being processed
				continue                       // Not enough to be complete
			}

			// We add the oprs, and the graded blocks. The next iteration of this loop will use these graded oprs.
			g.oprblks = append(g.oprblks, &OprBlock{
				Dbht:       block.EntryBlock.Header.DBHeight,
				OPRs:       oprs,
				GradedOPRs: graded,
			})

			// Let's add the winner's rewards. They will be happy that we do this step :)
			for place, winner := range graded[:10] { // Only top 10 matter
				reward := GetRewardFromPlace(place)
				if reward > 0 {
					err := AddToBalance(winner.CoinbasePNTAddress, reward)
					if err != nil {
						log.WithError(err).Fatal("Failed to update balance")
					}
					// Debug logs were here before to print the winners, it was a bit noisy
				}
			}

			// Debug log that the block was graded
			log.WithFields(log.Fields{
				"dbht":   block.EntryBlock.Header.DBHeight,
				"graded": len(oprs),
			}).Debugf("block graded")
		}
	}

	return nil
}

type OPRWorkRequest struct {
	entryhash string
}

type OPRWorkResponse struct {
	opr *OraclePriceRecord
	err error
}

// ParallelFetchOPRsFromEBlock is so we can parallelize our factomd requests.
func (g *QuickGrader) ParallelFetchOPRsFromEBlock(block *EntryBlockMarker) ([]*OraclePriceRecord, error) {
	// Previous winners so we know if the opr is valid
	// The Winners() wrapper just handles the base case for us, where there is no winners
	prevWinners := g.Winners(len(g.oprblks) - 1)
	numThreads := 4

	work := make(chan *OPRWorkRequest, numThreads*2)
	collect := make(chan *OPRWorkResponse, numThreads*2)

	// 10 threads
	for i := 0; i < numThreads; i++ {
		go g.fetchOPRWorker(work, collect, prevWinners, block.EntryBlock.Header.DBHeight)
	}
	count := len(block.EntryBlock.EntryList)

	var oprs []*OraclePriceRecord
	var collectErr error
	go func() {
		// Collection routine
		for resp := range collect {
			if resp.err != nil {
				collectErr = resp.err
			}
			if resp.opr != nil {
				oprs = append(oprs, resp.opr)
			}
			count--
			if count == 0 {
				break
			}
		}
		close(work)
	}()

	for _, entryHash := range block.EntryBlock.EntryList {
		work <- &OPRWorkRequest{entryhash: entryHash.EntryHash}
	}

	return oprs, collectErr
}

func (g *QuickGrader) fetchOPRWorker(work chan *OPRWorkRequest, results chan *OPRWorkResponse, prevWinners []*OraclePriceRecord, dbht int64) {
	for {
		select {
		case job, ok := <-work:
			if !ok {
				return // Done working
			}

			entry, err := factom.GetEntry(job.entryhash)
			if err != nil {
				results <- &OPRWorkResponse{err: fmt.Errorf("entry %s : %s", job.entryhash, err.Error())}
				continue
			}
			// If the opr is nil, the entry is not an opr. If the err is not nil, then something went wrong
			// that we need to retry. So the sync failed
			opr, err := g.ParseOPREntry(entry, dbht)
			if err != nil {
				results <- &OPRWorkResponse{err: err}
				continue
			}
			if opr == nil {
				results <- &OPRWorkResponse{opr: nil} // This entry is not correctly formatted
				continue
			}
			if !VerifyWinners(opr, prevWinners) {
				results <- &OPRWorkResponse{opr: nil} // This entry does not have the correct previous winners
				continue
			}

			results <- &OPRWorkResponse{opr: opr}
		}
	}
}

func (g *QuickGrader) FetchOPRsFromEBlock(block *EntryBlockMarker) ([]*OraclePriceRecord, error) {
	// Previous winners so we know if the opr is valid
	// The Winners() wrapper just handles the base case for us, where there is no winners
	prevWinners := g.Winners(len(g.oprblks) - 1)

	var oprs []*OraclePriceRecord
	for _, entryHash := range block.EntryBlock.EntryList {
		entry, err := factom.GetEntry(entryHash.EntryHash)
		if err != nil {
			return nil, fmt.Errorf("entry %s : %s", entryHash.EntryHash, err.Error())
		}
		// If the opr is nil, the entry is not an opr. If the err is not nil, then something went wrong
		// that we need to retry. So the sync failed
		opr, err := g.ParseOPREntry(entry, block.EntryBlock.Header.DBHeight)
		if err != nil {
			return nil, err
		}
		if opr == nil {
			continue // This entry is not correctly formatted
		}
		if !VerifyWinners(opr, prevWinners) {
			continue // This entry does not have the correct previous winners
		}

		oprs = append(oprs, opr)
	}
	return oprs, nil
}

// GetPreviousOPRBlock returns the winners of the previous OPR block
func (g *QuickGrader) GetPreviousOPRBlock(dbht int32) *OprBlock {
	for i := len(g.oprblks) - 1; i >= 0; i-- {
		if g.oprblks[i].Dbht < int64(dbht) {
			return g.oprblks[i]
		}
	}
	return nil
}

// GetPreviousOPRs returns the OPRs in highest-known block less than dbht.
// Returns nil if the dbht is the first dbht in the chain.
func (g *QuickGrader) GetPreviousOPRs(dbht int32) []*OraclePriceRecord {
	block := g.GetPreviousOPRBlock(dbht)
	if block != nil {
		return block.OPRs
	}
	return nil
}

func (g *QuickGrader) Winners(index int) (winners []*OraclePriceRecord) {
	if index == -1 {
		return winners // empty array is the base case
	}

	return g.oprblks[index].GradedOPRs[:10]
}

// ParseOPREntry will return the oracle price record for a given entry.
// 	Returns:
//		(opr, nil)		Entry is OPR and no errors
//		(nil, nil)		Entry is not an OPR and no errors
//		(nil, error)	We don't know, we should not check this eblock as processed
func (g *QuickGrader) ParseOPREntry(entry *factom.Entry, height int64) (*OraclePriceRecord, error) {
	var err error
	// Do some quick collecting of data and checks of the entry.
	// Can only have three ExtIDs which must be:
	//	[0] the nonce for the entry
	//	[1] Self reported difficulty
	//  [2] Version number
	if len(entry.ExtIDs) != 3 {
		return nil, nil
	}

	// Okay, it looks sort of okay.  Lets unmarshal the JSON
	opr := NewOraclePriceRecord()
	if err := json.Unmarshal(entry.Content, opr); err != nil {
		return nil, nil // Doesn't unmarshal, then it isn't valid for sure.  Continue on.
	}
	if opr.CoinbasePNTAddress, err = common.ConvertFCTtoPegNetAsset(g.Network, "PNT", opr.CoinbaseAddress); err != nil {
		return nil, nil // Invalid Coinbase Address
	}
	// Run some basic checks on the values.  If they don't check out, then ignore the entry
	if !opr.Validate(g.Config, height) {
		return nil, nil
	}

	// Keep this entry
	opr.EntryHash = entry.Hash()
	opr.Nonce = entry.ExtIDs[0]
	opr.SelfReportedDifficulty = entry.ExtIDs[1]
	if len(entry.ExtIDs[2]) != 1 {
		return nil, nil // Version is 1 byte
	}
	opr.Version = entry.ExtIDs[2][0]

	// Looking good.  Go ahead and compute the OPRHash
	sha := sha256.Sum256(entry.Content)
	opr.OPRHash = sha[:] // Save the OPRHash

	return opr, nil
}

func (g *QuickGrader) SendToListeners(winners *OPRs) {
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

/// -----

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
