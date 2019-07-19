// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package opr

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/FactomProject/factom"
	"github.com/dustin/go-humanize"
	"github.com/pegnet/pegnet/common"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

// Avg computes the average answer for the price of each token reported
func Avg(list []*OraclePriceRecord) (avg [20]float64) {

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
func CalculateGrade(avg [20]float64, opr *OraclePriceRecord) float64 {
	tokens := opr.GetTokens()
	opr.Grade = 0
	for i, v := range tokens {
		if avg[i] > 0 {
			d := (v.value - avg[i]) / avg[i] // compute the difference from the average
			opr.Grade = opr.Grade + d*d*d*d  // the grade is the sum of the squares of the differences
		} else {
			opr.Grade = v.value // If the average is zero, then it's all zero so
		} // set the Grade to the value.  It is as good a choice as any.
	}
	return opr.Grade
}

// GradeBlock takes all OPRs in a block and sorts them according to Grade and Difficulty.
// The top ten entries are considered the winners.
func GradeBlock(list []*OraclePriceRecord) (tobepaid []*OraclePriceRecord, sortedlist []*OraclePriceRecord) {

	list = RemoveDuplicateMiningIDs(list)

	if len(list) < 10 {
		return nil, nil
	}

	// Make sure we have the difficulty calculated for all items in the list.
	for _, v := range list {
		v.Difficulty = v.ComputeDifficulty(v.Entry.ExtIDs[0])
	}

	// Throw away all the entries but the top 50 on pure difficulty alone.
	// Note that we are sorting in descending order.
	sort.SliceStable(list, func(i, j int) bool { return list[i].Difficulty > list[j].Difficulty })

	if len(list) > 50 {
		list = list[:50]
	}
	for i := len(list); i >= 10; i-- {
		avg := Avg(list[:i])
		for j := 0; j < i; j++ {
			CalculateGrade(avg, list[j])
		}
		// Because this process can scramble the sorted fields, we have to resort with each pass.
		sort.SliceStable(list[:i], func(i, j int) bool { return list[i].Difficulty > list[j].Difficulty })
		sort.SliceStable(list[:i], func(i, j int) bool { return list[i].Grade < list[j].Grade })
	}
	tobepaid = append(tobepaid, list[:10]...)

	return tobepaid, list
}

// RemoveDuplicateMiningIDs runs a two-pass filter on the list to remove any duplicate entries.
// The entry with higher difficulty is kept.
// Two passes are used to avoid slice deletion logic
func RemoveDuplicateMiningIDs(list []*OraclePriceRecord) (nlist []*OraclePriceRecord) {
	// miner id => slice index of highest difficulty entry
	highest := make(map[string]int)

	for i, v := range list {
		id := strings.Join(v.FactomDigitalID, "-")

		if dupe, ok := highest[id]; ok { // look for duplicates
			if v.Difficulty <= list[dupe].Difficulty { // less then, we ignore
				continue
			}
		}
		// Either the first record found for the identity,or a more difficult record... keep it
		highest[id] = i

	}
	// Take all the best records, stick them in the list and return.
	for _, idx := range highest {
		nlist = append(nlist, list[idx])
	}
	return nlist
}

// block data at a specific height
type OprBlock struct {
	OPRs []*OraclePriceRecord
	Dbht int64
}

// OPRBlocks holds all the known OPRs
var OPRBlocks []*OprBlock

var ebMutex sync.Mutex

// GetEntryBlocks creates the OPR Records at a given dbht
func GetEntryBlocks(config *config.Config) {
	ebMutex.Lock()
	defer ebMutex.Unlock()

	p, err := config.String("Miner.Protocol")
	check(err)
	n, err := config.String("Miner.Network")
	check(err)
	opr := [][]byte{[]byte(p), []byte(n), []byte("Oracle Price Records")}
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
				// Can only have one ExtID which must be the nonce for the entry
				if len(entry.ExtIDs) != 1 {
					continue // keep looking if the entry has more than one extid
				}

				// Okay, it looks sort of okay.  Lets unmarshal the JSON
				opr := new(OraclePriceRecord)
				if err := json.Unmarshal(entry.Content, opr); err != nil {
					continue // Doesn't unmarshal, then it isn't valid for sure.  Continue on.
				}

				// Run some basic checks on the values.  If they don't check out, then ignore the entry
				if !opr.Validate(config) {
					continue
				}
				// Keep this entry
				opr.Entry = entry

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
				opr.Difficulty = opr.ComputeDifficulty(opr.Entry.ExtIDs[0])
			}
			validOPRs = append(validOPRs, opr) // Add to my valid list if all the winners are right
		}
		if len(validOPRs) < 10 { // Make sure we have at least 10 valid OPRs,
			continue // and leave if we don't.
		}
		winners, _ := GradeBlock(validOPRs)
		oprblocks[i].OPRs = winners
		OPRBlocks = append(OPRBlocks, oprblocks[i])

		if i == 0 {
			log.WithFields(log.Fields{
				"height": humanize.Comma(oprblocks[i].Dbht),
			}).Info("Added new valid block to OPR Chain")
		}

		// Update the balances for each winner
		for place, win := range winners {
			switch place {
			// The Big Winner
			case 0:
				err := AddToBalance(win.CoinbasePNTAddress, 800)
				if err != nil {
					log.WithError(err).Fatal("Failed to update balance")
				}
			// Second Place
			case 1:
				err := AddToBalance(win.CoinbasePNTAddress, 600)
				if err != nil {
					log.WithError(err).Fatal("Failed to update balance")
				}
			default:
				err := AddToBalance(win.CoinbasePNTAddress, 450)
				if err != nil {
					log.WithError(err).Fatal("Failed to update balance")
				}
			}
			fid := win.FactomDigitalID[0]
			for _, f := range win.FactomDigitalID[1:] {
				fid = fid + "-" + f
			}
			if i == 0 {
				log.WithFields(log.Fields{
					"place":      place,
					"fid":        fid,
					"entry_hash": hex.EncodeToString(win.Entry.Hash()[:8]),
					"grade":      fmt.Sprintf("%.4e", win.Grade),
					"difficulty": fmt.Sprintf("%.10e", float64(win.Difficulty)),
					"address":    win.CoinbasePNTAddress,
					"balance":    humanize.Comma(GetBalance(win.CoinbasePNTAddress)),
				}).Info("New OPR Winner")
			}
		}
	}
	return
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
