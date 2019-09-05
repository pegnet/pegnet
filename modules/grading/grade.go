package grading

import (
	"encoding/binary"
	"fmt"
	"math"
	"sort"

	"github.com/pegnet/pegnet/modules/opr"
	log "github.com/sirupsen/logrus"
)

const (
	// 1%
	GradeBand float64 = 0.01
)

// Avg computes the average answer for the price of each token reported
//	The list has to be in sorted in difficulty order before calling this function
func Avg(list []*opr.OPR) (avg []float64) {
	avg = make([]float64, len(list[0].GetTokens()))

	// Sum up all the prices
	for _, singleOpr := range list {
		tokens := singleOpr.GetTokens()
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
func CalculateGrade(avg []float64, opr *opr.OPR, band float64) float64 {
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
func (g *GradingBlock) GradeMinimum() (graded []*opr.OPR) {
	// No grade algo can handle 0
	if len(g.OPRs) == 0 {
		return nil
	}

	switch g.Version() {
	case 1:
		return g.gradeMinimumVersionOne()
	case 2:
		return g.gradeMinimumVersionTwo()
	}
	panic("Grading version unspecified")
}

// gradeMinimumVersionOne is the version 1 grading algorithm
// 1. PoW to top 50
// 2. Grading to top 10
// 3. Pay top 10 according to their place
func (g *GradingBlock) gradeMinimumVersionOne() (graded []*opr.OPR) {
	top50 := g.honestTop50(10)
	if top50 == nil {
		return nil
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

// gradeMinimumVersionTwo is version 2 grading algo
// 1. PoW to top 50
// 2. Grade with 1% tolerance band to top 25
// 3. Pay top 25 equally (not done here)
// 4. Grade to 1 without any tolerance band
// 5. Wining price is the last one
func (g *GradingBlock) gradeMinimumVersionTwo() (graded []*opr.OPR) {
	top50 := g.honestTop50(25)
	if top50 == nil {
		return nil
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

// honestTop50 goes through the oprs sorted by self reported difficulty, tossing those that
// are dishonest until we get 50. The `want` param allows us to short circuit if we have less than that,
// as the lxrhash is slow.
func (g *GradingBlock) honestTop50(want int) []*opr.OPR {
	// Sort the OPRs by self reported difficulty
	// We will toss dishonest ones as we walk down the list
	sort.SliceStable(g.OPRs, func(i, j int) bool {
		return binary.BigEndian.Uint64(g.OPRs[i].SelfReportedDifficulty) > binary.BigEndian.Uint64(g.OPRs[j].SelfReportedDifficulty)
	})

	list := RemoveDuplicateSubmissions(g.OPRs) // A copied list
	if len(list) < want {
		return nil
	}

	// Find the top 50 with the correct difficulties
	// 1. top50 is the top 50 PoW
	top50 := make([]*opr.OPR, 0)
	for _, singleOpr := range list {
		// Until we hit 50 records, we keep looking for honest records
		if !g.IsValidOPR(singleOpr, g.Height()) {
			continue
		}
		// Honest record found
		top50 = append(top50, singleOpr)
		if len(top50) == 50 {
			break // We have enough to grade
		}
	}
	return top50
}

// IsValidOPR will fully validate an opr. It will ensure the self reported difficulty is correct,
// and it's fields are set correctly. It will also validate the opr version matches the grading version.
func (g *GradingBlock) IsValidOPR(singleOpr *opr.OPR, dbht int32) bool {
	oprLog := g.Logger.WithFields(log.Fields{
		"entryhash": fmt.Sprintf("%x", singleOpr.EntryHash),
		"id":        singleOpr.FactomDigitalID,
		"dbtht":     singleOpr.Dbht,
	})

	if singleOpr.Version != g.Version() {
		oprLog.Warnf("running version %d. expected %d", singleOpr.Version, g.Version())
		return false
	}

	// Validation occurs here
	//	Check the previous winners
	if !VerifyWinners(singleOpr, g.PreviousWinners) {
		oprLog.Warnf("bad previous winners in opr")
		return false
	}

	//	Validate the fields
	if err := singleOpr.Validate(dbht); err != nil {
		oprLog.WithError(err).Warnf("opr failed to validate")
		return false
	}

	//	Check the self reported difficulty
	singleOpr.Difficulty = opr.ComputeDifficulty(singleOpr.OPRHash, singleOpr.Nonce)
	f := binary.BigEndian.Uint64(singleOpr.SelfReportedDifficulty)
	if f != singleOpr.Difficulty {
		oprLog.Warnf("Self reported difficulty incorrect Exp %x, found %x", singleOpr.Difficulty, singleOpr.SelfReportedDifficulty)
		return false
	}

	return true
}

// RemoveDuplicateSubmissions filters out any duplicate OPR (same nonce and OPRHash)
func RemoveDuplicateSubmissions(list []*opr.OPR) []*opr.OPR {
	// nonce+oprhash => exists
	added := make(map[string]bool)
	nlist := make([]*opr.OPR, 0)
	for _, v := range list {
		id := string(append(v.Nonce, v.OPRHash...))
		if !added[id] {
			nlist = append(nlist, v)
			added[id] = true
		}
	}
	return nlist
}

// VerifyWinners takes an opr and compares its list of winners to the winners of previousHeight
func VerifyWinners(opr *opr.OPR, winners []string) bool {
	// Is the list the same?
	for i, w := range opr.WinPreviousOPR {
		// If the winners is nil, then the opr.WinPrevious should all be blank
		if winners == nil && w != "" {
			return false
		}
		if len(winners) > 0 && w != winners[i] { // short hash
			return false
		}
	}
	return true
}
