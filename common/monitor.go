package common

// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import (
	"sync"
	"time"

	"github.com/FactomProject/factom"
)

// Timeout is the amount of time the Factomd endpoint can be unreachable
// before shutting down.
// This allows restarting the endpoint or short network outages.
var Timeout = time.Second * 90
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

// Monitor polls a factomd node and sends alerts whenever the block height or minute changes
type Monitor struct {
	polls int64

	listenerMutex sync.Mutex
	listeners     []chan MonitorEvent // Channels to send minutes to
	errorMutex    sync.Mutex
	errors        []chan error
}

// MonitorEvent is the data sent to all listeners when being notified
type MonitorEvent struct {
	Minute int64
	Dbht   int32
}

// Listener spawns a new listening channel that will receive updates of height or minute changes
func (f *Monitor) Listener() <-chan MonitorEvent {
	f.listenerMutex.Lock()
	defer f.listenerMutex.Unlock()

	listener := make(chan MonitorEvent, 10)
	f.listeners = append(f.listeners, listener)
	return listener
}

// ErrorListener spawns a new listening channel that will receive errors that have occurred
func (f *Monitor) ErrorListener() <-chan error {
	f.errorMutex.Lock()
	defer f.errorMutex.Unlock()

	listener := make(chan error, 1)
	f.errors = append(f.errors, listener)
	return listener
}

// waitForNextHeight polls the node every second until a new height is reached.
func (f *Monitor) waitForNextMinute(current factom.CurrentMinuteInfo) (factom.CurrentMinuteInfo, error) {
	end := time.Now().Add(Timeout)
	for {
		f.polls++
		info, err := factom.GetCurrentMinute()

		if err == nil {
			if info.DirectoryBlockHeight > current.DirectoryBlockHeight {
				return *info, nil
			}

			if info.DirectoryBlockHeight == current.DirectoryBlockHeight && current.Minute != info.Minute {
				return *info, nil
			}

			// the API has a lower height than the one we've seen
			time.Sleep(time.Second)
			continue
		}

		if end.Before(time.Now()) {
			return factom.CurrentMinuteInfo{}, err
		}
		time.Sleep(time.Millisecond * 100)
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
