package mining_test

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/pegnet/pegnet/opr"

	. "github.com/pegnet/pegnet/mining"
)

var totalIter = 100
var totalBytes = 40

func BenchmarkHash(b *testing.B) {
	b.Run("Sha256", benchmarkSha256)
	b.Run("LXR", benchmarkLXR)
}

func benchmarkSha256(b *testing.B) {
	data := make([]byte, totalBytes)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < totalIter; i++ {
			sha256.Sum256(data)
		}
	}
}

func benchmarkLXR(b *testing.B) {
	data := make([]byte, totalBytes)
	opr.InitLX()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < totalIter; i++ {
			opr.LX.Hash(data)
		}
	}
}

func TestMinerAction(t *testing.T) {
	var _ = t
}

// BenchmarkNonceRotate/simple_nonce_increment-8         	2000000000	         0.00 ns/op
func BenchmarkNonceRotate(b *testing.B) {
	b.Run("simple Nonce increment", testIncrement)
}

func TestRunitFast(t *testing.T) {
	var _ = t
	data := make([]byte, totalBytes)
	opr.InitLX()

	total := 10000
	n := time.Now()
	for i := 0; i < total; i++ {
		for i := 0; i < totalIter; i++ {
			opr.LX.Hash(data)
		}
	}

	fmt.Println(time.Since(n).Nanoseconds() / int64(total))
}

func testIncrement(b *testing.B) {
	ni := NewNonceIncrementer(1)
	for i := 0; i < b.N; i++ {
		ni.NextNonce()
	}
}

//func TestNonce(t *testing.T) {
//	ni := NewNonceIncrementer(1)
//	fmt.Println(int64(1) << 16)
//	for i := int64(0); i < int64(1)<<17; i++ {
//		ni.NextNonce()
//		//if i%255 == 0 {
//		fmt.Println(len(ni.Nonce[1:]), fmt.Sprintf("%x", ni.Nonce[1:]))
//		//}
//		//if len(ni.Nonce) > 3 {
//		//	t.Fatalf("wrong length: i: %d, i/255: %d, l: %d, n: %x", i, (i/255)+1, len(ni.Nonce), ni.Nonce)
//		//}
//
//	}
//}
