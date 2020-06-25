// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package spr

// SPRs is the message sent by the Grader
type SPRs struct {
	ToBePaid   []*StakingPriceRecord
	GradedOPRs []*StakingPriceRecord

	// Since this is used as a message, we need a way to send an error
	Error error
}
