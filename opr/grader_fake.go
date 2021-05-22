// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package opr

import (
	"github.com/pegnet/pegnet/balances"
	"github.com/pegnet/pegnet/database"
	"github.com/zpatrick/go-config"
)

// FakeGrader can be used in unit tests
type FakeGrader struct {
	*QuickGrader
}

func NewFakeGrader(config *config.Config, balances *balances.BalanceTracker) *FakeGrader {
	f := new(FakeGrader)
	f.QuickGrader = NewQuickGrader(config, database.NewMapDb(), balances)

	return f
}

func (f *FakeGrader) EmitFakeEvent(event OPRs) {
	for _, a := range f.alerts {
		select { // Don't block if someone isn't pulling from the winner channel
		case a <- &event:
		default:
			// This means the channel is full
		}
	}
}
