// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package opr

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"sort"

	"github.com/pegnet/pegnet/common"
	log "github.com/sirupsen/logrus"
)

const (
	// 1%
	GradeBand float64 = 0.01
)

// Avg computes the average answer for the price of each token reported
//	The list has to be in sorted in difficulty order before calling this function
func Avg(list []*OraclePriceRecord) (avg []float64) {
	avg = make([]float64, len(list[0].GetTokens()))

	// Sum up all the prices
	for _, opr := range list {
		tokens := opr.GetTokens()
		for i, token := range tokens {
			if token.Value >= 0 { // Make sure no OPR has negative values for
				avg[i] += token.Value // assets.  Simply treat all values as positive.
			} else {
				avg[i] -= token.Value
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
func CalculateGrade(avg []float64, opr *OraclePriceRecord, band float64) float64 {
	tokens := opr.GetTokens()
	opr.Grade = 0
	for i, v := range tokens {
		if avg[i] > 0 {
			d := (v.Value - avg[i]) / avg[i] // compute the difference from the average
			if band > 0 {
				d = ApplyBand(d, band)
			}
			opr.Grade = opr.Grade + d*d*d*d // the grade is the sum of the square of the square of the differences
		}
	}
	return opr.Grade
}

// ApplyBand
func ApplyBand(diff float64, band float64) float64 {
	diff = math.Abs(diff)
	// If the diff is less than the band, then our diff goes to 0
	if diff <= band {
		return 0
	}
	return diff - band
}

// GradeMinimum only grades the top 50 honest records. The input must be the records sorted by
// self reported difficulty.
func GradeMinimum(orderedList []*OraclePriceRecord, network string, dbht int64) (graded []*OraclePriceRecord) {
	// No grade algo can handle 0
	if len(orderedList) == 0 {
		return nil
	}

	switch common.OPRVersion(network, dbht) {
	case 1:
		return gradeMinimumVersionOne(orderedList)
	case 2, 3, 4, 5:
		return gradeMinimumVersionTwo(orderedList)
	}
	panic("Grading version unspecified")
}

// gradeMinimumVersionTwo is version 2 grading algo
// 1. PoW to top 50
// 2. Grade with 1% tolerance band to top 25
// 3. Pay top 25 equally (not done here)
// 4. Grade to 1 without any tolerance band
// 5. Wining price is the last one
func gradeMinimumVersionTwo(orderedList []*OraclePriceRecord) (graded []*OraclePriceRecord) {
	list := RemoveDuplicateSubmissions(orderedList)
	if len(list) < 25 {
		return nil
	}

	// Sort the OPRs by self reported difficulty
	// We will toss dishonest ones when we grade.
	sort.SliceStable(list, func(i, j int) bool {
		return binary.BigEndian.Uint64(list[i].SelfReportedDifficulty) > binary.BigEndian.Uint64(list[j].SelfReportedDifficulty)
	})

	// Find the top 50 with the correct difficulties
	// 1. top50 is the top 50 PoW
	top50 := make([]*OraclePriceRecord, 0)
	for i, opr := range list {
		opr.Difficulty = opr.ComputeDifficulty(opr.Nonce)
		f := binary.BigEndian.Uint64(opr.SelfReportedDifficulty)
		if f != opr.Difficulty {
			log.WithFields(log.Fields{
				"place":     i,
				"entryhash": fmt.Sprintf("%x", opr.EntryHash),
				"id":        opr.FactomDigitalID,
				"dbtht":     opr.Dbht,
			}).Warnf("Self reported difficulty incorrect Exp %x, found %x", opr.Difficulty, opr.SelfReportedDifficulty)
			continue
		}
		// Honest record
		top50 = append(top50, opr)
		if len(top50) == 50 {
			break // We have enough to grade
		}
	}

	// 2. Grade with 1% tolerance Band to top 25
	// 3. Pay top 25 (does not happen here)
	// 4. Grade to 1 without any tolerance band
	for i := len(top50); i >= 1; i-- {
		avg := Avg(top50[:i])
		for j := 0; j < i; j++ {
			band := 0.0
			if i >= 25 { // Use the band until we hit the 25
				band = GradeBand
			}
			CalculateGrade(avg, top50[j], band)
		}

		// Because this process can scramble the sorted fields, we have to resort with each pass.
		sort.SliceStable(top50[:i], func(i, j int) bool { return top50[i].Difficulty > top50[j].Difficulty })
		sort.SliceStable(top50[:i], func(i, j int) bool { return top50[i].Grade < top50[j].Grade })
	}
	return top50
}

// gradeMinimumVersionOne is the version 1 grading algorithm
// 1. PoW to top 50
// 2. Grading to top 10
// 3. Pay top 10 according to their place
func gradeMinimumVersionOne(orderedList []*OraclePriceRecord) (graded []*OraclePriceRecord) {
	list := RemoveDuplicateSubmissions(orderedList)
	if len(list) < 10 {
		return nil
	}

	// Sort the OPRs by self reported difficulty
	// We will toss dishonest ones when we grade.
	sort.SliceStable(list, func(i, j int) bool {
		return binary.BigEndian.Uint64(list[i].SelfReportedDifficulty) > binary.BigEndian.Uint64(list[j].SelfReportedDifficulty)
	})

	// Find the top 50 with the correct difficulties
	top50 := make([]*OraclePriceRecord, 0)
	for i, opr := range list {
		opr.Difficulty = opr.ComputeDifficulty(opr.Nonce)
		f := binary.BigEndian.Uint64(opr.SelfReportedDifficulty)
		if f != opr.Difficulty {
			log.WithFields(log.Fields{
				"place":     i,
				"entryhash": fmt.Sprintf("%x", opr.EntryHash),
				"id":        opr.FactomDigitalID,
				"dbht":      opr.Dbht,
			}).Warnf("Self reported difficulty incorrect Exp %x, found %x", opr.Difficulty, opr.SelfReportedDifficulty)
			continue
		}
		// Honest record
		top50 = append(top50, opr)
		if len(top50) == 50 {
			break // We have enough to grade
		}
	}

	for i := len(top50); i >= 10; i-- {
		avg := Avg(top50[:i])
		for j := 0; j < i; j++ {
			CalculateGrade(avg, top50[j], 0)
		}
		// Because this process can scramble the sorted fields, we have to resort with each pass.
		sort.SliceStable(top50[:i], func(i, j int) bool { return top50[i].Difficulty > top50[j].Difficulty })
		sort.SliceStable(top50[:i], func(i, j int) bool { return top50[i].Grade < top50[j].Grade })
	}
	return top50
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
	OPRs               []*OraclePriceRecord
	GradedOPRs         []*OraclePriceRecord
	Dbht               int64
	TotalNumberRecords int  // We tend to only keep the top 50, it's nice to know how many existed before we cut it
	EmptyOPRBlock      bool // An empty opr block is an eblock that could not totally validate
}

// VerifyWinners takes an opr and compares its list of winners to the winners of previousHeight
func VerifyWinners(opr *OraclePriceRecord, winners []*OraclePriceRecord) bool {
	for i, w := range opr.WinPreviousOPR {
		if winners == nil && w != "" {
			return false
		}
		if len(winners) > 0 && w != hex.EncodeToString(winners[i].EntryHash[:8]) { // short hash
			return false
		}
	}
	return true
}

func GetRewardFromPlace(place int, network string, height int64) int64 {
	switch common.OPRVersion(network, height) {
	case 1:
		return getRewardFromPlaceVersionOne(place)
	case 2, 3, 4, 5:
		return getRewardFromPlaceVersionTwo(place)
	}
	panic("opr version not found")
}

// getRewardFromPlaceVersionTwo pays top 25 evenly
func getRewardFromPlaceVersionTwo(place int) int64 {
	if place >= 25 {
		return 0 // There's no participation trophy. Return zero.
	}
	return 200 * 1e8
}

func getRewardFromPlaceVersionOne(place int) int64 {
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
