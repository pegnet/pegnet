// TODO: Move this to the LXRHash repo. All packages that import lxr30 should use the same global
package lxr30

import (
	"os"
	"strconv"
	"sync"

	lxr "github.com/pegnet/LXRHash"
)

// LX holds an instance of lxrhashh 30 bits for the hashmap
var LX30 lxr.LXRHash
var lxInitializer sync.Once

// The init function for LX is expensive. So we should explicitly call the init if we intend
// to use it. Make the init call idempotent
func InitLX() {
	lxInitializer.Do(func() {
		// This code will only be executed ONCE, no matter how often you call it
		LX30.Verbose(true)
		if size, err := strconv.Atoi(os.Getenv("LXRBITSIZE")); err == nil && size >= 8 && size <= 30 {
			LX30.Init(0xfafaececfafaecec, uint64(size), 256, 5)
		} else {
			LX30.Init(0xfafaececfafaecec, 30, 256, 5)
		}
	})
}

func Hash(src []byte) []byte {
	InitLX()
	return LX30.Hash(src)
}
