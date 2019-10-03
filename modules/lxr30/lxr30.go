// TODO: Move this to the LXRHash repo. All packages that import lxr30 should use the same global
package lxr30

import (
	"os"
	"strconv"
	"sync"

	lxr "github.com/pegnet/LXRHash"
)

var lx30 *lxr.LXRHash
var lxInitializer sync.Once

// The init function for LX is expensive. So we should explicitly call the init if we intend
// to use it. Make the init call idempotent
func InitLX() {
	lxInitializer.Do(func() {
		lx30 = new(lxr.LXRHash)
		// This code will only be executed ONCE, no matter how often you call it
		lx30.Verbose(true)
		if size, err := strconv.Atoi(os.Getenv("LXRBITSIZE")); err == nil && size >= 8 && size <= 30 {
			lx30.Init(0xfafaececfafaecec, uint64(size), 256, 5)
		} else {
			lx30.Init(0xfafaececfafaecec, 30, 256, 5)
		}
	})
}

func Init() *lxr.LXRHash {
	InitLX()
	return lx30
}
