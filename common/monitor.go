// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package common

import (
	"sync"
	"time"

	"github.com/FactomProject/factom"
	"github.com/cenkalti/backoff"
)

var monitor *Monitor
var once sync.Once

// GetMonitor returns the singleton instance of the Factomd Monitor
func GetMonitor() *Monitor {
	once.Do(func() {
		monitor = new(Monitor)
		monitor.errors = []chan error{}
		monitor.listeners = []chan MonitorEvent{}
		go monitor.poll()
	})
	return monitor
}

type IMonitor interface {
	NewListener() <-chan MonitorEvent
	NewErrorListener() <-chan error
	SetTimeout(timeout time.Duration)
}

// Monitor polls a factomd node and sends alerts whenever the block height or minute changes
type Monitor struct {
	timeout time.Duration

	listenerMutex sync.Mutex
	listeners     []chan MonitorEvent // Channels to send minutes to
	errorMutex    sync.Mutex
	errors        []chan error
}

// MonitorEvent is the Data sent to all listeners when being notified
type MonitorEvent struct {
	Minute int64 `json:"minute"`
	Dbht   int32 `json:"dbht"`
}

// NewListener spawns a new listening channel that will receive updates of height or minute changes
func (f *Monitor) NewListener() <-chan MonitorEvent {
	f.listenerMutex.Lock()
	defer f.listenerMutex.Unlock()

	listener := make(chan MonitorEvent, 10)
	f.listeners = append(f.listeners, listener)
	return listener
}

// NewErrorListener spawns a new listening channel that will receive errors that have occurred
func (f *Monitor) NewErrorListener() <-chan error {
	f.errorMutex.Lock()
	defer f.errorMutex.Unlock()

	listener := make(chan error, 1)
	f.errors = append(f.errors, listener)
	return listener
}

// SetTimeout sets a new timeout duration.
// If the monitor is unable to connect to the factomd node within the duration,
// an error is sent to all error listeners.
func (f *Monitor) SetTimeout(timeout time.Duration) {
	f.timeout = timeout
}

// getMinute queries the factomd api using exponential backoff.
func (f *Monitor) getMinute() (*factom.CurrentMinuteInfo, error) {
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = time.Millisecond * 100
	bo.MaxElapsedTime = f.timeout
	bo.MaxInterval = time.Second * 10
	bo.Reset()

	var info *factom.CurrentMinuteInfo
	var err error
	retry := func() error {
		info, err = factom.GetCurrentMinute()
		return err
	}

	fail := backoff.Retry(retry, bo)
	if fail != nil {
		return nil, fail
	}
	return info, nil
}

// waitForNextHeight polls the node every second until a new height is reached.
func (f *Monitor) waitForNextMinute(current factom.CurrentMinuteInfo) (factom.CurrentMinuteInfo, error) {
	for {
		info, err := f.getMinute()
		if err != nil {
			return factom.CurrentMinuteInfo{}, err
		}

		info.DirectoryBlockHeight += 1 // Add one so that the height represents the block being built

		if info.DirectoryBlockHeight > current.DirectoryBlockHeight {
			return *info, nil
		}

		if info.DirectoryBlockHeight == current.DirectoryBlockHeight && current.Minute != info.Minute {
			return *info, nil
		}

		// the API has a lower height than the one we've seen
		time.Sleep(time.Second)
	}
}

func (f *Monitor) notifyError(err error) {
	f.errorMutex.Lock()
	defer f.errorMutex.Unlock()
	for _, ec := range f.errors {
		select {
		case ec <- err:
		default:
		}
	}
}

func (f *Monitor) notify(info factom.CurrentMinuteInfo) {
	f.listenerMutex.Lock()
	defer f.listenerMutex.Unlock()

	fds := MonitorEvent{
		Dbht:   int32(info.DirectoryBlockHeight),
		Minute: info.Minute,
	}
	for _, l := range f.listeners {
		select {
		case l <- fds:
		default:
		}
	}
}

// poll the factomd node and notify listeners of minute/height changes
func (f *Monitor) poll() {
	var info factom.CurrentMinuteInfo
	var err error
	for {
		info, err = f.waitForNextMinute(info)
		if err != nil {
			f.notifyError(err)
			continue
		}
		f.notify(info)
	}
}
