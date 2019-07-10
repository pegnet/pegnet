package polling

// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import (
	"time"

	"github.com/cenkalti/backoff"
	log "github.com/sirupsen/logrus"
)

// Default values for PollingExponentialBackOff.
const (
	DefaultInitialInterval     = 800 * time.Millisecond
	DefaultRandomizationFactor = 0.5
	DefaultMultiplier          = 1.5
	DefaultMaxInterval         = 6 * time.Second
	DefaultMaxElapsedTime      = 30 * time.Second // max 60 seconds
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


func ConverToUnix(format string, value string) (timestamp int64) {
	t, err := time.Parse(format, value)
	if err != nil {
		log.WithError(err).Fatal("Failed to convert timestamp")
	}
	return t.Unix()
}
