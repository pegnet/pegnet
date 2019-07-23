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
