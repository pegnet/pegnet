package opr

import (
	"encoding/hex"
	"encoding/json"
	"github.com/FactomProject/btcutil/base58"
	"github.com/FactomProject/factom"
	"github.com/pegnet/OracleRecord/support"
	"github.com/zpatrick/go-config"
	"sync"
)

// Compute the average answer for the price of each token reported
func Avg(list []*OraclePriceRecord) (avg [20]float64) {
	// Sum up all the prices
	for _, opr := range list {
		tokens := opr.GetTokens()
		for i, price := range tokens {
			avg[i] += price.value
		}
	}
	// Then divide the prices by the number of OraclePriceRecord records.  Two steps is actually faster
	// than doing everything in one loop (one divide for every asset rather than one divide
	// for every asset * number of OraclePriceRecords)
	numList := float64(len(list))
	for i := range avg {
		avg[i] = avg[i] / numList / 100000000
	}
	return
}

// Given the average answers across a set of tokens, grade the opr
func CalculateGrade(avg [20]float64, opr *OraclePriceRecord) float64 {
	tokens := opr.GetTokens()
	for i, v := range tokens {
		d := v.value - avg[i]           // compute the difference from the average
		opr.Grade = opr.Grade + d*d*d*d // the grade is the sum of the squares of the differences
	}
	return opr.Grade
}

// Given a list of OraclePriceRecord, figure out which 10 should be paid, and in what order
func GradeBlock(list []*OraclePriceRecord) (tobepaid []*OraclePriceRecord, sortedlist []*OraclePriceRecord) {

	if len(list) <= 10 {
		return nil, nil
	}

	last := len(list)
	// Throw away all the entries but the top 50 in difficulty
	// bubble sort because I am lazy.  Could be replaced with about anything
	for j := 0; j < len(list)-1; j++ {
		for k := 0; k < len(list)-j-1; k++ {
			d1 := list[k].Difficulty
			d2 := list[k+1].Difficulty
			if d1 == 0 || d2 == 0 {
				//panic("Should not be here")
			}
			if d1 < d2 { // sort the smallest difficulty to the end of the list
				list[k], list[k+1] = list[k+1], list[k]
			}
		}
	}
	if len(list) > 50 {
		last = 50
	}
	// Go through and throw away entries that are outside the average or on a tie, have the worst difficulty
	// until we are only left with 10 entries to reward
	for i := last; i >= 10; i-- {
		avg := Avg(list[:i])
		for j := 0; j < i; j++ {
			CalculateGrade(avg, list[j])
		}
		// bubble sort the worst grade to the end of the list. Note that this is nearly sorted data, so
		// a bubble sort with a short circuit is pretty darn good sort.
		for j := 0; j < i-1; j++ {
			cont := false                // If we can get through a pass with no swaps, we are done.
			for k := 0; k < i-j-1; k++ { // yes, yes I know we can get 2 or 3 x better speed playing with indexes
				if list[k].Grade > list[k+1].Grade { // bit it is tricky.  This is good enough.
					list[k], list[k+1] = list[k+1], list[k] // sort first by the grade.
					cont = true                             // any swap means we continue to loop
				} else if list[k].Grade == list[k+1].Grade { // break ties with PoW.  Where data is being shared
					if list[k].Difficulty < list[k+1].Difficulty { // we will have ties.
						//list[k], list[k+1] = list[k+1], list[k]
						cont = true // any swap means we continue to loop
					}
				}
			}
			if !cont { // If we made a pass without any swaps, we are done.
				break
			}
		}
	}
	tobepaid = append(tobepaid, list[:10]...)
	return tobepaid, list
}

var EntryBlocks []*factom.EBlock
var Entries map[string]*factom.Entry
var EBMutex sync.Mutex

// Get the OPR Records at a given dbht
func GetEntryBlocks(config *config.Config) {
	EBMutex.Lock()
	defer EBMutex.Unlock()
	if Entries == nil {
		Entries = make(map[string]*factom.Entry, 100)
	}

	var entryBlocks []*factom.EBlock

	p, err := config.String("Miner.Protocol")
	check(err)
	n, err := config.String("Miner.Network")
	check(err)
	opr := [][]byte{[]byte(p), []byte(n), []byte("Oracle Price Records")}
	heb, err := factom.GetChainHead(hex.EncodeToString(support.ComputeChainIDFromFields(opr)))
	check(err)
	eb, err := factom.GetEBlock(heb)
	check(err)
	elen := len(EntryBlocks)
	for eb != nil && (elen == 0 || eb.Header.DBHeight > EntryBlocks[elen-1].Header.DBHeight) {
		entryBlocks = append(entryBlocks, eb)
		for _, ebentry := range eb.EntryList {
			entry, err := factom.GetEntry(ebentry.EntryHash)
			check(err)

			// All OPRs have one and only 1 external ID
			if len(entry.ExtIDs) == 1 {
				Entries[ebentry.EntryHash] = entry
			}

		}
		neb, err := factom.GetEBlock(eb.Header.PrevKeyMR)
		if err != nil {
			break
		}
		eb = neb
	}
	for i := len(entryBlocks) - 1; i >= 0; i-- {
		EntryBlocks = append(EntryBlocks, entryBlocks[i])
	}
	return
}

func GetPreviousOPRs(dbht int32) []*OraclePriceRecord {
	EBMutex.Lock()
	defer EBMutex.Unlock()

	eblen := len(EntryBlocks)
	for i := eblen - 1; i >= 0; i-- {
		if EntryBlocks[i].Header.DBHeight < int64(dbht) {
			oprs := GetOPRs(EntryBlocks[i])
			if oprs != nil {
				return oprs
			}
		}
	}
	return nil
}

func GetOPRs(eblock *factom.EBlock) (oprs []*OraclePriceRecord) {
	for _, ebentry := range eblock.EntryList {
		if Entries[ebentry.EntryHash] != nil {
			opr := new(OraclePriceRecord)

			opr.Entry = Entries[ebentry.EntryHash]
			if opr.Entry == nil {
				continue
			}
			// Decode the EntryHash to its binary form.  If that doesn't work, this entry isn't good.
			eh, err1 := hex.DecodeString(ebentry.EntryHash)
			// Unmarshal the entry, and add to our list if it works.
			err2 := json.Unmarshal(opr.Entry.Content, &opr)
			if err1 != nil || err2 != nil { // If it doesn't unmarshal, then just ignore
				delete(Entries, ebentry.EntryHash) // could be trash someone put in our chain
				continue
			}
			oprs = append(oprs, opr)

			// Compute the OPRHash here because we need it to compute difficulty
			opr.OPRHash = LX.Hash(opr.Entry.Content)
			// Keep the EntryHash with the entry because we need it to build the next OPR record
			opr.EntryHash = base58.Encode(eh)

			diff := opr.ComputeDifficulty(opr.Entry.ExtIDs[0])
			opr.Difficulty = diff

		}
	}
	return
}
