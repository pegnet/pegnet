// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package common

import (
	"time"

	"github.com/cenkalti/backoff"
)

// Default values for PegExponentialBackOff.
const (
	DefaultInitialInterval     = 500 * time.Millisecond
	DefaultRandomizationFactor = 0.5
	DefaultMultiplier          = 1.5
	DefaultMaxInterval         = 2 * time.Second
	DefaultMaxElapsedTime      = 10 * time.Second // max 10 seconds
)

// PegExponentialBackOff creates an instance of ExponentialBackOff
func PegExponentialBackOff() *backoff.ExponentialBackOff {
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
