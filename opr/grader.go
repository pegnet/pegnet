// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package opr

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/pegnet/pegnet/balances"

	"github.com/pegnet/pegnet/database"

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
			winners.ToBePaid = oprs.GradedOPRs[:10]
			winners.AllOPRs = oprs.OPRs

			// Alert followers that we have graded the previous block
			g.SendToListeners(&winners)

			// TODO: This should be another routine, not affecting grading
			err = g.Burns.UpdateBurns(g.Config, g.GetFirstOPRBlock().Dbht)
			if err != nil {
				log.WithField("id", "grader").WithError(err).Errorf("error processing burns")
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
			for place, winner := range oprblock.GradedOPRs[:10] { // Only top 10 matter
				reward := GetRewardFromPlace(place)
				if reward > 0 {
					err := g.Balances.AddToBalance(winner.CoinbasePNTAddress, reward)
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
	// There is not enough entries in this block, so there is no point in looking at it.
	if len(block.EntryBlock.EntryList) < 10 {
		return nil, nil
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
		return nil, err
	}

	// Check if we have enough oprs for this block. If we have less than 10, there is no point in
	// trying to grade it.
	if len(oprs) < 10 {
		return nil, nil // Not enough oprs for this block be a valid oprblock
	}

	// Sort the OPRs by self reported difficulty
	// We will toss dishonest ones when we grade.
	sort.SliceStable(oprs, func(i, j int) bool {
		return binary.BigEndian.Uint64(oprs[i].SelfReportedDifficulty) > binary.BigEndian.Uint64(oprs[j].SelfReportedDifficulty)
	})

	// GradeMinimum will only grade the first 50 honest records
	graded := GradeMinimum(oprs)
	if len(graded) < 10 { // We might lose some when we reject dishonest records
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
	g.oprBlkLock.Lock()
	prevWinners := g.Winners(len(g.oprBlks) - 1)
	g.oprBlkLock.Unlock()

	numThreads := 4

	work := make(chan *OPRWorkRequest, numThreads*2)
	collect := make(chan *OPRWorkResponse, numThreads*2)

	// 10 threads
	for i := 0; i < numThreads; i++ {
		go g.fetchOPRWorker(work, collect, prevWinners, block.EntryBlock.Header.DBHeight)
	}
	count := len(block.EntryBlock.EntryList)

	var wg sync.WaitGroup
	wg.Add(count)

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
			wg.Done()
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

	wg.Wait()

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
	g.oprBlkLock.Lock()
	prevWinners := g.Winners(len(g.oprBlks) - 1)
	g.oprBlkLock.Unlock()

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

func (g *QuickGrader) Winners(index int) (winners []*OraclePriceRecord) {
	if index == -1 {
		return winners // empty array is the base case
	}

	return g.oprBlks[index].GradedOPRs[:10]
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
	if len(entry.ExtIDs[1]) != 8 { // self reported difficulty must be a uint64
		return nil, nil
	}
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

// oprByHash returns the entire OPR based on it's hash
func (g *QuickGrader) OprByHash(hash string) OraclePriceRecord {
	g.oprBlkLock.Lock()
	defer g.oprBlkLock.Unlock()

	for _, block := range g.oprBlks {
		for _, record := range block.OPRs {
			if hash == hex.EncodeToString(record.OPRHash) {
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
			if bytes.Compare(hashBytes, record.OPRHash[:8]) == 0 {
				return *record
			}
		}
	}
	return OraclePriceRecord{}
}

/// -----

// OPRs is the message sent by the Grader
type OPRs struct {
	ToBePaid []*OraclePriceRecord
	AllOPRs  []*OraclePriceRecord

	// Since this is used as a message, we need a way to send an error
	Error error
}
