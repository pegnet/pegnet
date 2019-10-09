// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package polling

import (
	"math"
	"math/rand"
	"time"

	"github.com/pegnet/pegnet/common"
)

const qlimit = 580 // Limit queries to once just shy of 10 minutes (600 seconds)

type PegAssets map[string]PegItem

func (p PegAssets) Clone(randomize float64) PegAssets {
	np := make(PegAssets)
	for _, asset := range common.AllAssets {
		np[asset] = p[asset].Clone(randomize)
	}

	return np
}

type PegItem struct {
	Value    uint64
	WhenUnix int64 // unix timestamp
	When     time.Time
}

func Uint64Value(value float64) uint64 {
	return uint64(math.Round(value * 1e8))
}

// Value FloatValue the value to a float
// Deprecated: Should not be using floats, stick to the uint64
func (p PegItem) FloatValue() float64 {
	return float64(p.Value) / 1e8
}

func (p PegItem) Clone(randomize float64) PegItem {
	np := new(PegItem)
	np.Value = Uint64Value(p.FloatValue() + p.FloatValue()*(randomize/2*rand.Float64()) - p.FloatValue()*(randomize/2*rand.Float64()))
	np.WhenUnix = p.WhenUnix
	np.When = p.When
	return *np
}
