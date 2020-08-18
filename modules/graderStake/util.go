package graderStake

import (
	"encoding/hex"
)

// DecodeError indicates that there was a problem Unmarshalling the content
type DecodeError struct{ Msg string }

func (d *DecodeError) Error() string {
	return d.Msg
}

// NewDecodeError creates a new DecodeError with the specified message
func NewDecodeError(m string) *DecodeError {
	return &DecodeError{Msg: m}
}

// ValidateError indicates that there is an incompatibility with the data
type ValidateError struct{ Msg string }

func (v *ValidateError) Error() string {
	return v.Msg
}

// NewValidateError creates a new ValidateError with the specified message
func NewValidateError(m string) *ValidateError {
	return &ValidateError{Msg: m}
}

// ensures there are `length` winners and either they're all zero
// or they're all 8 byte hexadecimal
func verifyWinnerFormat(winners []string, length int) bool {
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

// returns true if both slices are identical
func verifyWinners(have []string, wanted []string) bool {
	if len(have) != len(wanted) {
		return false
	}
	for i := range have {
		if have[i] != wanted[i] {
			return false
		}
	}
	return true
}
