// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package polling

import (
	"math/rand"
	"time"
)

const qlimit = 580 // Limit queries to once just shy of 10 minutes (600 seconds)

type PegAssets map[string]PegItem

func (p PegAssets) Clone(randomize float64) PegAssets {
	np := make(PegAssets)
	for asset := range p {
		np[asset] = p[asset].Clone(randomize)
	}

	return np
}

type PegItem struct {
	Value    float64
	WhenUnix int64 // unix timestamp
	When     time.Time
}

func (p PegItem) Clone(randomize float64) PegItem {
	np := new(PegItem)
	np.Value = p.Value + p.Value*(randomize/2*rand.Float64()) - p.Value*(randomize/2*rand.Float64())
	np.Value = TruncateTo8(np.Value)
	np.WhenUnix = p.WhenUnix
	return *np
}
