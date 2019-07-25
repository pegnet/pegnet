package mining_test

import (
	"crypto/sha256"
	"testing"

	. "github.com/pegnet/pegnet/mining"
	"github.com/pegnet/pegnet/opr"
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
	data := make([]byte, 32)
	n := NewNonceIncrementer(0)
	opr.InitLX()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < totalIter; i++ {
			n.NextNonce()
			opr.LX.Hash(append(n.Nonce, data...))
		}
	}
}

// BenchmarkNonceRotate/simple_nonce_increment-8         	2000000000	         0.00 ns/op
func BenchmarkNonceRotate(b *testing.B) {
	b.Run("simple Nonce increment", testIncrement)
}

func testIncrement(b *testing.B) {
	ni := NewNonceIncrementer(1)
	for i := 0; i < b.N; i++ {
		ni.NextNonce()
	}
}
