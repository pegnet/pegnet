// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package common

// FakeMonitor can be used in unit tests
type FakeMonitor struct {
	*Monitor
}

func NewFakeMonitor() *FakeMonitor {
	f := new(FakeMonitor)
	monitor = new(Monitor)
	monitor.errors = []chan error{}
	monitor.listeners = []chan MonitorEvent{}
	f.Monitor = monitor

	return f
}

func (f *FakeMonitor) FakeNotify(dbht int, minute int) {
	f.listenerMutex.Lock()
	defer f.listenerMutex.Unlock()

	fds := MonitorEvent{
		Dbht:   int32(dbht),
		Minute: int64(minute),
	}

	for _, l := range f.listeners {
		select {
		case l <- fds:
		default:
		}
	}
}
