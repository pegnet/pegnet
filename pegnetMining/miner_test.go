package pegnetMining_test

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/opr"

	. "github.com/pegnet/pegnet/pegnetMining"
)

func TestMinerID(t *testing.T) {
	for i := 1; i < 10; i++ {
		if id := int(GetNextMinerID()); id != i {
			t.Errorf("Exp %d, got %d", i, id)
		}
	}
}

// TestMinerCleanup checks that stopping the miner actually stops all the threads
func TestMinerCleanup(t *testing.T) {
	g := opr.NewFakeGrader()
	m := common.NewFakeMonitor()

	p := NewPegnetMiner(common.NewUnitTestConfig(), m, g)
	t.Run("Before first event stop", func(t *testing.T) {
		n := runtime.NumGoroutine()

		go p.LaunchMiningThread(false)
		for n == runtime.NumGoroutine() { // Let thread start
			time.Sleep(1 * time.Millisecond)
		}
		// We need to let the channels get made
		time.Sleep(10 * time.Millisecond)

		n = runtime.NumGoroutine()
		if !p.StopMining() {
			t.Errorf("should be started")
		}

		start := time.Now()
		for n != runtime.NumGoroutine() { // Let thread start
			time.Sleep(1 * time.Millisecond)
			if time.Since(start) > 1*time.Second {
				t.Fatal("Miner did not close properly")
			}
		}
	})

	t.Run("Before first event stop", func(t *testing.T) {
		n := runtime.NumGoroutine()

		go p.LaunchMiningThread(false)
		for n == runtime.NumGoroutine() { // Let thread start
			time.Sleep(1 * time.Millisecond)
		}
		// We need to let the channels get made
		time.Sleep(10 * time.Millisecond)

		// Start the work for min 1
		m.FakeNotify(1, 1)
		time.Sleep(10 * time.Millisecond) // Let other thread go

		n = runtime.NumGoroutine()
		if !p.StopMining() {
			t.Errorf("should be started")
		}

		start := time.Now()
		for n != runtime.NumGoroutine() { // Let thread start
			time.Sleep(1 * time.Millisecond)
			if time.Since(start) > 1*time.Second {
				fmt.Println(n, runtime.NumGoroutine())
				t.Fatal("Miner did not close properly")
			}
		}
	})

	t.Run("Stop stopped miner", func(t *testing.T) {
		if p.StopMining() {
			t.Errorf("should be stopped")
		}
	})
}

func TestMinerOPR(t *testing.T) {

}
