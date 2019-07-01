package common_test

import (
	"fmt"
	. "github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/opr"
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	DoLogging = true
	done := 0
	go func() {
		for i := 0; i < 10; i++ {
			Logf("Every1", "Do something every second %d", i)
			time.Sleep(1 * time.Second)
		}
		done++
	}()

	go func() {
		for i := 0; i < 5; i++ {
			Logf("Every2", "Do something every two seconds %d", i)
			time.Sleep(2 * time.Second)
		}
		done++
	}()

	Do(func() {
		fmt.Println("Doing Hashes")
		hash := []byte{1, 2, 3}
		for i := 1; i < 500000; i++ {
			hash = opr.LX.Hash(hash)
		}
		Logf("PriceyHash", "%x", hash)
		done++
	})

	Logf("multiline", "%s", "line1\nline2\nline3\nline4")

	DoLogging = false
	Do(func() {
		t.Fail()
		hash := []byte{1, 2, 3}
		for i := 1; i < 500000; i++ {
			hash = opr.LX.Hash(hash)
		}
		Logf("PriceyHash", "%x", hash)
	})
	for done < 3 {
		time.Sleep(100 * time.Millisecond)
	}
}
