// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package opr

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"sort"
	"sync"

	"github.com/FactomProject/factom"
	"github.com/dustin/go-humanize"
	"github.com/pegnet/pegnet/common"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

// Avg computes the average answer for the price of each token reported
func Avg(list []*OraclePriceRecord) (avg []float64) {
	avg = make([]float64, len(common.AllAssets))

	// Sum up all the prices
	for _, opr := range list {
		tokens := opr.GetTokens()
		for i, token := range tokens {
			if token.value >= 0 { // Make sure no OPR has negative values for
				avg[i] += token.value // assets.  Simply treat all values as positive.
			} else {
				avg[i] -= token.value
			}
		}
	}
	// Then divide the prices by the number of OraclePriceRecord records.  Two steps is actually faster
	// than doing everything in one loop (one divide for every asset rather than one divide
	// for every asset * number of OraclePriceRecords)  There is also a little bit of a precision advantage
	// with the two loops (fewer divisions usually does help with precision) but that isn't likely to be
	// interesting here.
	numList := float64(len(list))
	for i := range avg {
		avg[i] = avg[i] / numList
	}
	return
}

// CalculateGrade takes the averages and grades the individual OPRs
func CalculateGrade(avg []float64, opr *OraclePriceRecord) float64 {
	tokens := opr.GetTokens()
	opr.Grade = 0
	for i, v := range tokens {
		if avg[i] > 0 {
			d := (v.value - avg[i]) / avg[i] // compute the difference from the average
			opr.Grade = opr.Grade + d*d*d*d  // the grade is the sum of the square of the square of the differences
		}
	}
	return opr.Grade
}

// GradeBlock takes all OPRs in a block, sorts them according to Difficulty, and grades the top 50.
// The top ten graded entries are considered the winners. Returns the top 50 sorted by grade, then the original list
// sorted by difficulty.
func GradeBlock(list []*OraclePriceRecord) (graded []*OraclePriceRecord, sorted []*OraclePriceRecord) {
	list = RemoveDuplicateSubmissions(list)

	if len(list) < 10 {
		return nil, nil
	}

	// Throw away all the entries but the top 50 on pure difficulty alone.
	// Note that we are sorting in descending order.
	sort.SliceStable(list, func(i, j int) bool { return list[i].Difficulty > list[j].Difficulty })

	var topDifficulty []*OraclePriceRecord
	if len(list) > 50 {
		topDifficulty = make([]*OraclePriceRecord, 50)
		copy(topDifficulty[:50], list[:50])
	} else {
		topDifficulty = make([]*OraclePriceRecord, len(list))
		copy(topDifficulty, list)
	}
	for i := len(topDifficulty); i >= 10; i-- {
		avg := Avg(topDifficulty[:i])
		for j := 0; j < i; j++ {
			CalculateGrade(avg, topDifficulty[j])
		}
		// Because this process can scramble the sorted fields, we have to resort with each pass.
		sort.SliceStable(topDifficulty[:i], func(i, j int) bool { return topDifficulty[i].Difficulty > topDifficulty[j].Difficulty })
		sort.SliceStable(topDifficulty[:i], func(i, j int) bool { return topDifficulty[i].Grade < topDifficulty[j].Grade })
	}
	return topDifficulty, list // Return the top50 sorted by grade and then all sorted by difficulty
}

// RemoveDuplicateSubmissions filters out any duplicate OPR (same nonce and OPRHash)
func RemoveDuplicateSubmissions(list []*OraclePriceRecord) []*OraclePriceRecord {
	// nonce+oprhash => exists
	added := make(map[string]bool)
	nlist := make([]*OraclePriceRecord, 0)
	for _, v := range list {
		id := string(append(v.Nonce, v.OPRHash...))
		if !added[id] {
			nlist = append(nlist, v)
			added[id] = true
		}
	}
	return nlist
}

// block data at a specific height
type OprBlock struct {
	OPRs       []*OraclePriceRecord
	GradedOPRs []*OraclePriceRecord
	Dbht       int64
}

// OPRBlocks holds all the known OPRs
var OPRBlocks []*OprBlock

var ebMutex sync.Mutex

// GetEntryBlocks creates the OPR Records at a given dbht
func GetEntryBlocks(config *config.Config) {
	ebMutex.Lock()
	defer UpdateBurns(config)
	defer ebMutex.Unlock()

	network, err := common.LoadConfigNetwork(config)
	check(err)
	p, err := config.String("Miner.Protocol")
	check(err)
	n, err := common.LoadConfigNetwork(config)
	check(err)
	opr := [][]byte{[]byte(p), []byte(n), []byte(common.OPRChainTag)}
	heb, _, err := factom.GetChainHead(hex.EncodeToString(common.ComputeChainIDFromFields(opr)))
	check(err)
	eb, err := factom.GetEBlock(heb)
	check(err)

	// A temp list of candidate oprblocks to evaluate to see if they fit nicely together
	// Because we go from the head of the chain backwards to collect them, they have to be
	// collected before I can then validate them forward from the highest valid OPR block
	// I have found.
	var oprblocks []*OprBlock
	// For each entryblock in the Oracle Price Records chain
	// Get all the valid OPRs and put them in  a new OPRBlock structure
	for eb != nil && (len(OPRBlocks) == 0 ||
		eb.Header.DBHeight > OPRBlocks[len(OPRBlocks)-1].Dbht) {

		// Go through the Entry Block and collect all the valid OPR records
		if len(eb.EntryList) > 10 {
			oprblk := new(OprBlock)
			oprblk.Dbht = eb.Header.DBHeight
			for _, ebentry := range eb.EntryList {
				entry, err := factom.GetEntry(ebentry.EntryHash)
				check(err)

				// Do some quick collecting of data and checks of the entry.
				// Can only have two ExtIDs which must be:
				//	[0] the nonce for the entry
				//	[1] Self reported difficulty
				if len(entry.ExtIDs) != 2 {
					continue // keep looking if the entry has more than one extid
				}

				// Okay, it looks sort of okay.  Lets unmarshal the JSON
				opr := NewOraclePriceRecord()
				if err := json.Unmarshal(entry.Content, opr); err != nil {
					continue // Doesn't unmarshal, then it isn't valid for sure.  Continue on.
				}
				if opr.CoinbasePNTAddress, err = common.ConvertFCTtoPegNetAsset(network, "PNT", opr.CoinbaseAddress); err != nil {
					continue
				}
				// Run some basic checks on the values.  If they don't check out, then ignore the entry
				if !opr.Validate(config, oprblk.Dbht) {
					continue
				}
				// Keep this entry
				opr.EntryHash = entry.Hash()
				opr.Nonce = entry.ExtIDs[0]
				opr.SelfReportedDifficulty = entry.ExtIDs[1]

				// Looking good.  Go ahead and compute the OPRHash
				opr.OPRHash = LX.Hash(entry.Content) // Save the OPRHash

				// Okay, mostly good.  Add to our candidate list
				oprblk.OPRs = append(oprblk.OPRs, opr)

			}
			// If we have 10 canidates, then lets add them up.
			if len(oprblk.OPRs) >= 10 {
				oprblocks = append(oprblocks, oprblk)
			}
		}
		// At this point, the oprblk has all the valid OPRs. Make sure we have enough.
		// sorted list of winners.

		neb, err := factom.GetEBlock(eb.Header.PrevKeyMR)
		if err != nil {
			break
		}
		eb = neb
	}

	// Take the reverse ordered opr blocks, from last to first.  Validate all the winners are
	// the right winners.  Replace the generally correct OPR list in the opr block with the
	// list of winners.  These should be the winners of the next block, which lucky enough is
	// the next block we are going to process.
	// Ignore opr blocks that don't get 10 winners.
	for i := len(oprblocks) - 1; i >= 0; i-- { // Okay, go through these backwards
		prevOPRBlock := GetPreviousOPRs(int32(oprblocks[i].Dbht)) // Get the previous OPRBlock
		var validOPRs []*OraclePriceRecord                        // Collect the valid OPRPriceRecords here
		for _, opr := range oprblocks[i].OPRs {                   // Go through this block
			for j, eh := range opr.WinPreviousOPR { // Make sure the winning records are valid
				if (prevOPRBlock == nil && eh != "") ||
					(prevOPRBlock != nil && eh != prevOPRBlock[0].WinPreviousOPR[j]) {
					continue
				}
				opr.Difficulty = opr.ComputeDifficulty(opr.Nonce)

				// TODO: Enforce this?
				f := binary.BigEndian.Uint64(opr.SelfReportedDifficulty)
				if f != opr.Difficulty {
					//This is a falsely reported difficulty. There is nothing we can
					//really do. Maybe we should log.warn how many per block are 'malicious'?
					log.Errorf("Diff mistmatch. Exp %d, found %d", opr.Difficulty, f)
				}

			}
			validOPRs = append(validOPRs, opr) // Add to my valid list if all the winners are right
		}
		if len(validOPRs) < 10 { // Make sure we have at least 10 valid OPRs,
			continue // and leave if we don't.
		}
		gradedOPRs, sortedOPRs := GradeBlock(validOPRs)
		oprblocks[i].GradedOPRs = gradedOPRs
		oprblocks[i].OPRs = sortedOPRs
		OPRBlocks = append(OPRBlocks, oprblocks[i])

		// Update the balances for each winner
		for place, winner := range gradedOPRs[:10] {
			reward := GetRewardFromPlace(place)
			if reward > 0 {
				err := AddToBalance(winner.CoinbasePNTAddress, reward)
				if err != nil {
					log.WithError(err).Fatal("Failed to update balance")
				}
			}
			if i == 0 {
				logger := log.WithFields(log.Fields{
					"place":      place,
					"id":         winner.FactomDigitalID,
					"entry_hash": hex.EncodeToString(winner.EntryHash[:8]),
					"grade":      common.FormatGrade(winner.Grade, 4),
					"difficulty": common.FormatDiff(winner.Difficulty, 10),
					"address":    winner.CoinbasePNTAddress,
					"balance":    humanize.Comma(GetBalance(winner.CoinbasePNTAddress)),
				})
				if place == 0 {
					logger.Info("New OPR Winner")
				} else {
					logger.Debug("New OPR Winner")
				}
			}
		}
	}
}

// GetPreviousOPRs returns the OPRs in highest-known block less than dbht.
// Returns nil if the dbht is the first dbht in the chain.
func GetPreviousOPRs(dbht int32) []*OraclePriceRecord {
	for i := len(OPRBlocks) - 1; i >= 0; i-- {
		if OPRBlocks[i].Dbht < int64(dbht) {
			return OPRBlocks[i].OPRs
		}
	}
	return nil
}

func GetRewardFromPlace(place int) int64 {
	if place >= 10 {
		return 0 // There's no participation trophy. Return zero.
	}
	switch place {
	case 0:
		return 800 * 1e8 // The Big Winner
	case 1:
		return 600 * 1e8 // Second Place
	default:
		return 450 * 1e8 // Consolation Prize
	}
}
