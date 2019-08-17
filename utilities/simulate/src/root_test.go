package src_test

import (
	"testing"

	"github.com/pegnet/pegnet/utilities/simulate/src"
)

func BenchmarkRandomDifficulty(b *testing.B) {
	for i := 0; i < b.N; i++ {
		src.RandomDifficulty()
	}
}
