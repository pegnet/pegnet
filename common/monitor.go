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
var monitor *FactomdMonitor
var once sync.Once

// GetMonitor returns the singleton instance of the Factomd Monitor
func GetMonitor() *FactomdMonitor {
	once.Do(func() {
		monitor = new(FactomdMonitor)
		monitor.listeners = []chan FDStatus{}
		go monitor.poll()
	})
	return monitor
}

// FactomdMonitor polls a factomd node and sends alerts whenever the block height or minute changes
type FactomdMonitor struct {
	lastMinute    int64 // Last minute we got
	lastBlock     int64 // Last block we got
	listenerMutex sync.Mutex
	listeners     []chan FDStatus // Channels to send minutes to
	polls         int64
	info          *factom.CurrentMinuteInfo
	status        string
}

// Listener spawns a new listening channel that will receive updates of height or minute changes
func (f *FactomdMonitor) Listener() <-chan FDStatus {
	f.listenerMutex.Lock()
	defer f.listenerMutex.Unlock()

	listener := make(chan FDStatus, 10)
	f.listeners = append(f.listeners, listener)
	return listener
}

// waitForNextHeight polls the node every second until a different height is reached.
// If the underlying factomd is swapped out this might result in a time rollback
func (f *FactomdMonitor) waitForNextMinute(limit int) *factom.CurrentMinuteInfo {
	end := time.Now().Add(Timeout)
	for {
		f.polls++
		info, err := factom.GetCurrentMinute()

		if err != nil {
			f.status = err.Error()
		} else if info.DirectoryBlockHeight < f.lastBlock {
			// restarting?
		} else if info.DirectoryBlockHeight > f.lastBlock || info.Minute != f.lastMinute {
			return info
		}

		if end.Before(time.Now()) {
			panic("Monitor: unable to retrieve current minute info: " + f.status)
		}
		time.Sleep(time.Second)
	}
}

// poll the factomd node and notify listeners of minute/height changes
func (f *FactomdMonitor) poll() {
	for {
		info := f.waitForNextMinute(90)
		f.info = info
		f.lastMinute = info.Minute
		f.lastBlock = info.DirectoryBlockHeight

		fds := FDStatus{
			Dbht:   int32(f.lastBlock),
			Minute: f.lastMinute,
		}

		// send alerts to all interested parties
		f.listenerMutex.Lock()
		for _, alert := range f.listeners {
			select {
			case alert <- fds:
			default:
			}
		}
		f.listenerMutex.Unlock()
	}
}
