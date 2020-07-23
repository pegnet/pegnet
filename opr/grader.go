// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package opr

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/FactomProject/factom"
	"github.com/pegnet/pegnet/balances"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/database"
	log "github.com/sirupsen/logrus"
	config "github.com/zpatrick/go-config"
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

	Balances *balances.BalanceTracker
	Burns    *balances.BurnTracking
	OPRChain *EntryBlockSync

	Config *config.Config

	// oprBlks is all the eblocks that contain the oprs
	oprBlks    []*OprBlock
	oprBlkLock sync.Mutex

	BlockStore IOPRBlockStore

	// lastGraded is the last graded oprblk, so we know
	// where to start grading
	lastGraded int

	alerts      map[string]chan *OPRs
	alertsMutex sync.Mutex // Maps are not thread safe
}

func NewQuickGrader(config *config.Config, db database.IDatabase, balanceTraker *balances.BalanceTracker) *QuickGrader {
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
	g.oprBlks = make([]*OprBlock, 0)

	g.BlockStore = NewOPRBlockStore(db)
	g.Balances = balanceTraker
	g.Burns = balances.NewBurnTracking(g.Balances)

	return g
}

func (g *QuickGrader) Close() error {
	log.Info("closing grader db")
	return g.BlockStore.Close()
}

// GetBlocks should only be used in unit tests. It is not thread safe
func (g *QuickGrader) GetBlocks() []*OprBlock {
	return g.oprBlks
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
	log.WithField("id", "grader").Info("Running initial sync")
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
			if oprs != nil && len(oprs.GradedOPRs) > 0 {
				// top 10 get paid for v1, top 25 for v2
				amt := g.MinRecords(int64(oprs.GradedOPRs[0].Dbht))
				// TODO: We should really only send the Graded, rather than graded and ToBePaid
				if len(oprs.GradedOPRs) >= amt {
					winners.ToBePaid = oprs.GradedOPRs[:amt]
				}
				winners.GradedOPRs = oprs.GradedOPRs
			}

			// Alert followers that we have graded the previous block
			g.SendToListeners(&winners)

			firstOPR := g.GetFirstOPRBlock()
			if firstOPR != nil {
				// TODO: This should be another routine, not affecting grading
				err = g.Burns.UpdateBurns(g.Config, firstOPR.Dbht)
				if err != nil {
					log.WithFields(log.Fields{
						"id": "grader",
					}).WithError(err).Errorf("error processing burns")
				}
			}
		}

	}
}

// Sync will sync our opr chain to the latest eblock head of the OPR chain
func (g *QuickGrader) Sync() error {
	fLog := log.WithField("id", "gradersync")

	// Syncblocks will take our chain and gather all the eblocks
	// that might remain to be synced. This means this function ONLY syncs eblocks
	// and from there we can sync the blocks one by one
	fLog.Debugf("syncing eblocks")
	err := g.OPRChain.SyncBlocks()
	if err != nil {
		return err
	}

	dbheight := int64(0)
	fLog.Debugf("syncing entries")
	startAmt := len(g.OPRChain.BlocksToBeParsed)
	// If we have eblocks to sync, this is where we go through them one by one.
	if !g.OPRChain.Synced() {
		c := 0
		// We have eblocks to sync!
		// NextEBlock() will return the next eblock in the chain that we still need to sync
		// it's entries. So we can walk, NextEblock -> NextEblock -> ... to sync the whole chain.
		for block := g.OPRChain.NextEBlock(); block != nil; block = g.OPRChain.NextEBlock() {
			dbheight = block.EntryBlock.Header.DBHeight
			c++
			done := startAmt - len(g.OPRChain.BlocksToBeParsed)
			if c%30 == 0 || done == startAmt {
				fLog.WithFields(log.Fields{
					"dbht": block.EntryBlock.Header.DBHeight,
				}).Debugf("syncing entries, %.2f%%", float64(done)/float64(startAmt)*100)
			}

			var err error
			// Before we try to fetch from the net, we try and fetch from disk
			oprblock, _ := g.BlockStore.FetchOPRBlock(block.EntryBlock.Header.DBHeight)
			if oprblock == nil {
				// Fetch from factomd
				oprblock, err = g.FetchOPRBlock(block)
			} else {
				if oprblock.EmptyOPRBlock {
					g.OPRChain.BlockParsed(*block)
					continue // This eblock does not have a valid opr block
				}
			}

			// If we have an error from factomd
			if err != nil {
				return err
			}

			if oprblock == nil {
				err := g.BlockStore.WriteInvalidOPRBlock(block.EntryBlock.Header.DBHeight)
				if err != nil {
					return err
				}
				g.OPRChain.BlockParsed(*block) // This block is done being processed
				continue
			}

			g.oprBlkLock.Lock()
			// We add the oprs, and the graded blocks. The next iteration of this loop will use these graded oprs.
			err = g.BlockStore.WriteOPRBlock(oprblock)
			if err != nil {
				g.oprBlkLock.Unlock()
				return err
			}
			g.oprBlks = append(g.oprBlks, oprblock)
			g.oprBlkLock.Unlock()

			// Let's add the winner's rewards. They will be happy that we do this step :)
			payouts := g.MinRecords(dbheight)
			for place, winner := range oprblock.GradedOPRs[:payouts] { // The top 25 matter in version 2
				reward := GetRewardFromPlace(place, g.Network, block.EntryBlock.Header.DBHeight)
				if reward > 0 {
					err := g.Balances.AddToBalance(winner.CoinbasePEGAddress, reward)
					if err != nil {
						log.WithError(err).Fatal("Failed to update balance")
					}
					// Debug logs were here before to print the winners, it was a bit noisy
				}
			}

			g.OPRChain.BlockParsed(*block)
		}
	}
	fLog.WithField("height", dbheight).Debugf("synced!")

	return nil
}

func (g *QuickGrader) FetchOPRBlock(block *EntryBlockMarker) (*OprBlock, error) {
	min := g.MinRecords(block.EntryBlock.Header.DBHeight)
	// There is not enough entries in this block, so there is no point in looking at it.
	if len(block.EntryBlock.EntryList) < min {
		return nil, nil
	}

	var oprs []*OraclePriceRecord
	var err error
	// Multithread when there is a lot (like a 6x speedup in my tests against mainnet)
	oprs, err = g.ParallelFetchOPRsFromEBlock(block, 4, true)

	if err != nil {
		return nil, err
	}

	// Check if we have enough oprs for this block. If we have less than 10, there is no point in
	// trying to grade it.
	if len(oprs) < min {
		return nil, nil // Not enough oprs for this block be a valid oprblock
	}

	// GradeMinimum will only grade the first 50 honest records
	graded := GradeMinimum(oprs, g.Network, block.EntryBlock.Header.DBHeight)
	if len(graded) < min { // We might lose some when we reject dishonest records
		return nil, nil // Not enough to be complete
	}

	// We are saving all of the oprs here, even the ones that were not graded.
	// TODO: Should we save them all? Or truncate here
	oprblock := &OprBlock{
		Dbht:               block.EntryBlock.Header.DBHeight,
		OPRs:               oprs,
		GradedOPRs:         graded,
		TotalNumberRecords: len(oprs),
	}
	return oprblock, nil
}

type OPRWorkRequest struct {
	order     int
	entryhash string
}

type OPRWorkResponse struct {
	order int // The original entry order
	opr   *OraclePriceRecord
	err   error
}

// ParallelFetchOPRsFromEBlock is so we can parallelize our factomd requests.
// 	Params:
//		block
//		workerCount		Number of parallel outbound requests
//		enforceWinners	Verify the previous winners
func (g *QuickGrader) ParallelFetchOPRsFromEBlock(block *EntryBlockMarker, workerCount int, enforceWinners bool) ([]*OraclePriceRecord, error) {
	// Previous winners so we know if the opr is valid
	// The Winners() wrapper just handles the base case for us, where there is no winners
	prevWinners := g.GetPreviousWinners(int32(block.EntryBlock.Header.DBHeight))

	// Using *2 just gives us a buffer so nothing is every blocking
	work := make(chan *OPRWorkRequest, workerCount*2)
	collect := make(chan *OPRWorkResponse, workerCount*2)

	// 10 threads
	for i := 0; i < workerCount; i++ {
		go g.fetchOPRWorker(work, collect, prevWinners, block.EntryBlock.Header.DBHeight, enforceWinners)
	}
	count := len(block.EntryBlock.EntryList)

	var wg sync.WaitGroup
	wg.Add(count)

	var oprResponses []*OPRWorkResponse
	var collectErr error
	go func() {
		// Collection routine
		for resp := range collect {
			if resp.err != nil {
				collectErr = resp.err
			}
			if resp.opr != nil {
				oprResponses = append(oprResponses, resp)
			}
			wg.Done()
			count--
			if count == 0 {
				break
			}
		}
		close(work)
	}()

	for i, entryHash := range block.EntryBlock.EntryList {
		work <- &OPRWorkRequest{entryhash: entryHash.EntryHash, order: i}
	}

	wg.Wait()

	sort.SliceStable(oprResponses, func(i, j int) bool { return oprResponses[i].order < oprResponses[j].order })
	// Now grab oprs
	oprs := make([]*OraclePriceRecord, len(oprResponses))
	for i := range oprResponses {
		oprs[i] = oprResponses[i].opr
	}

	return oprs, collectErr
}

func (g *QuickGrader) fetchOPRWorker(work chan *OPRWorkRequest, results chan *OPRWorkResponse, prevWinners []*OraclePriceRecord, dbht int64, enforceWinners bool) {
	for {
		select {
		case job, ok := <-work:
			if !ok {
				return // Done working
			}

			entry, err := factom.GetEntry(job.entryhash)
			if err != nil {
				results <- &OPRWorkResponse{err: fmt.Errorf("entry %s : %s", job.entryhash, err.Error()), order: job.order}
				continue
			}
			// If the opr is nil, the entry is not an opr. If the err is not nil, then something went wrong
			// that we need to retry. So the sync failed
			opr, err := g.ParseOPREntry(entry, dbht)
			if err != nil {
				results <- &OPRWorkResponse{err: err, order: job.order}
				continue
			}
			if opr == nil {
				results <- &OPRWorkResponse{opr: nil, order: job.order} // This entry is not correctly formatted
				continue
			}
			if enforceWinners && !VerifyWinners(opr, prevWinners) {
				log.WithFields(log.Fields{
					"entryhash": fmt.Sprintf("%x", opr.EntryHash),
					"id":        opr.FactomDigitalID,
					"dbht":      opr.Dbht,
				}).Warnf("bad previous winners in opr")
				results <- &OPRWorkResponse{opr: nil, order: job.order} // This entry does not have the correct previous winners
				continue
			}

			results <- &OPRWorkResponse{opr: opr, order: job.order}
		}
	}
}

// GetPreviousOPRBlock returns the winners of the previous OPR block
func (g *QuickGrader) GetPreviousOPRBlock(dbht int32) *OprBlock {
	g.oprBlkLock.Lock()
	defer g.oprBlkLock.Unlock()

	for i := len(g.oprBlks) - 1; i >= 0; i-- {
		if g.oprBlks[i].Dbht < int64(dbht) {
			return g.oprBlks[i]
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

func (g *QuickGrader) GetFirstOPRBlock() *OprBlock {
	g.oprBlkLock.Lock()
	defer g.oprBlkLock.Unlock()

	if len(g.oprBlks) == 0 {
		return nil
	}

	return g.oprBlks[0]
}

func (g *QuickGrader) GetPreviousWinners(dbht int32) (winners []*OraclePriceRecord) {
	prev := g.GetPreviousOPRBlock(dbht)
	if prev == nil {
		return winners // empty array is the base case
	}
	amt := g.MinRecords(prev.Dbht)

	return prev.GradedOPRs[:amt]
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

	// Need the version number
	if len(entry.ExtIDs[2]) != 1 {
		return nil, nil
	}
	opr.Version = entry.ExtIDs[2][0]

	if err := opr.SafeUnmarshal(entry.Content); err != nil {
		return nil, nil // Doesn't unmarshal, then it isn't valid for sure.  Continue on.
	}
	if opr.CoinbasePEGAddress, err = common.ConvertFCTtoPegNetAsset(g.Network, "PEG", opr.CoinbaseAddress); err != nil {
		return nil, nil // Invalid Coinbase Address
	}
	// Run some basic checks on the values.  If they don't check out, then ignore the entry
	if !opr.Validate(g.Config, height) {
		return nil, nil
	}

	// Keep this entry
	opr.EntryHash = entry.Hash()
	opr.Nonce = entry.ExtIDs[0]
	if len(entry.ExtIDs[1]) != 8 { // self reported difficulty must be a uint64
		return nil, nil
	}
	opr.SelfReportedDifficulty = entry.ExtIDs[1]

	// Looking good.  Go ahead and compute the OPRHash
	sha := sha256.Sum256(entry.Content)
	opr.OPRHash = sha[:] // Save the OPRHash

	// Set this information so we know what grading version to use
	opr.Network = g.Network
	opr.Protocol = g.Protocol

	return opr, nil
}

func (g *QuickGrader) MinRecords(dbht int64) int {
	switch common.OPRVersion(g.Network, dbht) {
	case 1:
		return 10
	case 2, 3, 4, 5:
		return 25
	default:
		panic("didn't get a valid opr version")
	}
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

// oprBlockByHeight returns a single OPRBlock
func (g *QuickGrader) OprBlockByHeight(dbht int64) *OprBlock {
	g.oprBlkLock.Lock()
	defer g.oprBlkLock.Unlock()

	for _, block := range g.oprBlks {
		if block.Dbht == dbht {
			return block
		}
	}
	return nil
}

// oprsByDigitalID returns every OPR created by a given ID
// Multiple ID's per miner or single daemon are possible.
// This function searches through every possible ID and returns all.
func (g *QuickGrader) OprsByDigitalID(did string) []OraclePriceRecord {
	g.oprBlkLock.Lock()
	defer g.oprBlkLock.Unlock()

	var subset []OraclePriceRecord
	for _, block := range g.oprBlks {
		for _, record := range block.OPRs {
			if record.FactomDigitalID == did {
				subset = append(subset, *record)
			}
		}
	}
	return subset
}

// oprByHash returns the entire OPR based on it's entry hash
func (g *QuickGrader) OprByHash(hash string) OraclePriceRecord {
	g.oprBlkLock.Lock()
	defer g.oprBlkLock.Unlock()

	for _, block := range g.oprBlks {
		for _, record := range block.OPRs {
			if hash == hex.EncodeToString(record.EntryHash) {
				return *record
			}
		}
	}
	return OraclePriceRecord{}
}

// Failing tests. Need to grok how the short 8 byte winning oprhashes are done.
func (g *QuickGrader) OprByShortHash(shorthash string) OraclePriceRecord {
	g.oprBlkLock.Lock()
	defer g.oprBlkLock.Unlock()

	hashBytes, _ := hex.DecodeString(shorthash)
	// hashbytes = reverseBytes(hashbytes)
	for _, block := range g.oprBlks {
		for _, record := range block.OPRs {
			if bytes.Compare(hashBytes, record.EntryHash[:8]) == 0 {
				return *record
			}
		}
	}
	return OraclePriceRecord{}
}

// OPRs is the message sent by the Grader
type OPRs struct {
	ToBePaid   []*OraclePriceRecord
	GradedOPRs []*OraclePriceRecord

	// Since this is used as a message, we need a way to send an error
	Error error
}

/// -----

// DEBUGAddOPRBlock is used for unit tests. We need access to the private field
// to setup some basic testing
func (g *QuickGrader) DEBUGAddOPRBlock(oprBlock *OprBlock) {
	g.oprBlks = append(g.oprBlks, oprBlock)
}
