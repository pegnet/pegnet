package mining_test

import (
	"testing"

	. "github.com/pegnet/pegnet/mining"
)

// BenchmarkNonceRotate/simple_nonce_increment-8         	2000000000	         0.00 ns/op
func BenchmarkNonceRotate(b *testing.B) {
	b.Run("simple Nonce increment", testIncrement)
}

func testIncrement(b *testing.B) {
	ni := NewNonceIncrementer(1)
	for i := 0; i < 255; i++ {
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
