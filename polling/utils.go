// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package polling

import (
	"math"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pegnet/pegnet/common"
	log "github.com/sirupsen/logrus"
)

// Default values for PollingExponentialBackOff.
const (
	DefaultInitialInterval     = 800 * time.Millisecond
	DefaultRandomizationFactor = 0.5
	DefaultMultiplier          = 1.5
	DefaultMaxInterval         = 6 * time.Second
	DefaultMaxElapsedTime      = 30 * time.Second // max 30 seconds
)

// PollingExponentialBackOff creates an instance of ExponentialBackOff
func PollingExponentialBackOff() *backoff.ExponentialBackOff {
	b := &backoff.ExponentialBackOff{
		InitialInterval:     DefaultInitialInterval,
		RandomizationFactor: DefaultRandomizationFactor,
		Multiplier:          DefaultMultiplier,
		MaxInterval:         DefaultMaxInterval,
		MaxElapsedTime:      DefaultMaxElapsedTime,
		Clock:               backoff.SystemClock,
	}
	b.Reset()
	return b
}

const precision = 10000

// RoundRate truncates the float
func RoundRate(v float64) float64 {
	return math.Round(v*precision) / precision
}

func ConverToUnix(format string, value string) (timestamp int64) {
	t, err := time.Parse(format, value)
	if err != nil {
		log.WithError(err).Fatal("Failed to convert timestamp")
	}
	return t.Unix()
}

func UpdatePegAssets(rates map[string]float64, timestamp int64, peg PegAssets, prefix ...string) {
	p := ""
	if len(prefix) > 0 {
		p = prefix[0]
	}

	for _, currencyISO := range common.CurrencyAssets {
		if v, ok := rates[p+currencyISO]; ok {
			peg[currencyISO] = PegItem{Value: v, When: timestamp}
		}
	}
}
