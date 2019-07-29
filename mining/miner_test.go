package mining_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
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
	data := make([]byte, 32)
	n := NewNonceIncrementer(0)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < totalIter; i++ {
			n.NextNonce()
			sha256.Sum256(append(n.Nonce, data...))
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

// BenchmarkNonceRotate/simple_nonce_increment-8         	200000000	         7.94 ns/op
func BenchmarkNonceRotate(b *testing.B) {
	b.Run("simple Nonce increment", testIncrement)
}

func testIncrement(b *testing.B) {
	ni := NewNonceIncrementer(1)
	for i := 0; i < b.N; i++ {
		ni.NextNonce()
	}
}

func TestNonceIncrementer(t *testing.T) {
	incrs := make([]*NonceIncrementer, 0)
	for i := 0; i < 256; i++ {
		incrs = append(incrs, NewNonceIncrementer(i))
	}
	used := make(map[string]bool)

	// convert []byte to int
	c := func(b []byte) int {
		var r int
		for i := 0; i < len(b); i++ {
			r <<= 8
			r += int(b[i])
		}
		return r
	}

	var a int
	for i := 0; i < 0x10000; i++ {
		a = c(incrs[0].Nonce[1:])

		if a != i {
			t.Fatalf("n1 mismatched i. want = %d, got = %d, raw = %s", i, a, hex.EncodeToString(incrs[0].Nonce))
		}

		for _, inc := range incrs {
			if bytes.Compare(incrs[0].Nonce[1:], inc.Nonce[1:]) != 0 {
				t.Fatalf("mismatch at %d. n0 = %s, n%d = %s", i, hex.EncodeToString(incrs[0].Nonce[1:]), inc.Nonce[0], hex.EncodeToString(inc.Nonce[1:]))
			}
		}

		for _, inc := range incrs {
			if used[string(inc.Nonce)] {
				t.Fatalf("nonce id%d %d already seen before", inc.Nonce[0], i)
			}
			used[string(inc.Nonce)] = true
		}

		for _, inc := range incrs {
			inc.NextNonce()
		}
	}
}
