package grader

import "encoding/hex"

// ensures there are `length` winners and either they're all zero
// or they're all 8 byte hexadecimal
func verifyWinners(winners []string, length int) bool {
	if len(winners) != length {
		return false
	}

	notEmpty := len(winners) > 0 && len(winners[0]) > 0

	for _, s := range winners {
		if notEmpty {
			if len(s) != 16 {
				return false
			}
			_, err := hex.DecodeString(s)
			if err != nil {
				return false
			}

		} else {
			if len(s) != 0 {
				return false
			}
		}
	}

	return true
}
