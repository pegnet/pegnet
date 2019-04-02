package oprecord

// Compute the average answer for the price of each token reported
func Avg(list []*OraclePriceRecord) (avg [20]float64) {
	// Sum up all the prices
	for _, opr := range list {
		tokens := opr.GetTokens()
		for i, price := range tokens {
			avg[i] += float64(price)
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
		d := float64(v)/100000000 - avg[i] // compute the difference from the average
		opr.Grade = opr.Grade + d*d*d*d    // the grade is the sum of the squares of the differences
	}
	return opr.Grade
}

// Given a list of OraclePriceRecord, figure out which 10 should be paid, and in what order
func GradeBlock(list []*OraclePriceRecord) (tobepaid []*OraclePriceRecord, sortedlist []*OraclePriceRecord) {

	if len(list) <= 10 {
		return nil, nil
	}

	// Calculate the difficult for each entry in the list of OraclePriceRecords.
	for _, opr := range list {
		opr.ComputeDifficulty()
	}

	last := len(list)
	// Throw away all the entries but the top 50 in difficulty
	if len(list) > 50 {
		// bubble sort because I am lazy.  Could be replaced with about anything
		for j := 0; j < len(list)-1; j++ {
			for k := 0; k < len(list)-j-1; k++ {
				d1 := list[k].Difficulty
				d2 := list[k+1].Difficulty
				if d1 == 0 || d2 == 0 {
					panic("Should not be here")
				}
				if d1 > d2 { // sort the largest difficulty to the end of the list
					list[k], list[k+1] = list[k+1], list[k]
				}
			}
		}
		last = 50 // truncate the list to the best 50
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
					if list[k].Difficulty > list[k+1].Difficulty { // we will have ties.
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
