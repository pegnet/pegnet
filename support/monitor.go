// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package support

import (
	"github.com/FactomProject/factom"
	"github.com/pegnet/pegnet/common"
	"math/rand"
	"sync"
	"time"
)

// FactomdMonitor
// Running multiple Monitors is problematic and should be avoided if possible
type FactomdMonitor struct {
	root       bool                   // True if this is the root FactomMonitor
	mutex      sync.Mutex             // Protect multiple parties accessing monitor data
	lastminute int64                  // Last minute we got
	lastblock  int64                  // Last block we got
	polltime   int64                  // How frequently do we poll
	kill       chan int               // Channel to kill polling.
	response   chan int               // Respond when we have stopped
	alerts     []chan common.FDStatus // Channels to send minutes to
	polls      int64
	info       *factom.CurrentMinuteInfo
	status     string
}

func (f *FactomdMonitor) GetAlert() chan common.FDStatus {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	alert := make(chan common.FDStatus, 10)
	f.alerts = append(f.alerts, alert)
	return alert
}

// GetBlockTime
// Returns the blocktime in seconds.  All blocks are divided into 10 "minute" sections.  But if the blocktime
// is not 600 seconds, then a minute = the blocktime/10
func (f *FactomdMonitor) GetBlockTime() int64 {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.info.DirectoryBlockInSeconds
}

// Returns the highest saved block
func (f *FactomdMonitor) GetHighestSavedDBlock() int64 {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.info.DirectoryBlockHeight
}

// poll
// Go process to poll the Factoid client to provide insight into its operations.
func (f *FactomdMonitor) poll() {
	for {
		var err error
		for {
			f.mutex.Lock()

			// If we have a kill message, die!
			select {
			case <-f.kill:
				f.response <- 1
				f.mutex.Unlock()
				return
			default:
			}

			for {
				var dbht int64
				if f.info != nil {
					dbht = f.info.DirectoryBlockHeight
				}
				// Do our poll
				f.info, err = factom.GetCurrentMinute()

				f.mutex.Unlock()

				for i := 0; i < 1000 && err != nil; i++ {
					time.Sleep(time.Duration(rand.Intn(50)+50) * time.Millisecond)
					// Do our poll
					f.mutex.Lock()
					f.info, err = factom.GetCurrentMinute()
					f.mutex.Unlock()

					if f.info != nil && err == nil {
						break
					}
				}
				f.mutex.Lock()
				if dbht < f.info.DirectoryBlockHeight { // Keep looking if the dbht hasn't progressed forward
					continue // Mostly this happens when the factomd node is rebooted.
				}
				f.info.DirectoryBlockHeight = dbht
				break
			}

			// track how often we poll
			f.polls++

			// If we get an error, then report and break
			if err != nil {
				f.status = err.Error()
				panic("Error with getting minute. " + f.status)

				break
			}
			// If we got a different block time, consider that good and break
			if f.info.Minute != f.lastminute || f.info.DirectoryBlockHeight != f.lastblock {
				f.lastminute = f.info.Minute
				f.lastblock = f.info.DirectoryBlockHeight
				break
			}

			// Poll once per second until we get a new minute
			f.mutex.Unlock()
			time.Sleep(1 * time.Second)
		}

		// send alerts to all interested parties
		for _, alert := range f.alerts {
			if cap(alert) > len(alert) {
				var fds common.FDStatus
				fds.Dbht = int32(f.info.DirectoryBlockHeight)
				fds.Minute = f.info.Minute
				alert <- fds
			}
		}

		f.mutex.Unlock()
		// Poll once per second
		time.Sleep(time.Duration(time.Second))
	}
}

func (f *FactomdMonitor) Start() {
	f.mutex.Lock()
	if f.kill == nil {
		f.response = make(chan int, 1)
		f.alerts = []chan common.FDStatus{}
		f.kill = make(chan int, 1)
		factom.SetFactomdServer("localhost:8088")
		factom.SetWalletServer("localhost:8089")
		go f.poll()
	}
	f.mutex.Unlock()
	return
}

func (f *FactomdMonitor) Stop() {
	f.mutex.Lock()
	kill := f.kill
	response := f.response
	f.mutex.Unlock()

	kill <- 0
	<-response
}
