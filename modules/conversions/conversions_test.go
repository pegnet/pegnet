package conversions_test

import (
	"testing"

	. "github.com/pegnet/pegnet/modules/conversions"
	"github.com/stretchr/testify/assert"
)

var conversionTests = []struct {
	Name           string
	Error          string
	FromAmount     int64
	FromRate       uint64
	ToRate         uint64
	ExpectedResult int64
}{{
	Name:           "invalid (zero fromRate)",
	Error:          "invalid rate: 0",
	FromAmount:     1,
	FromRate:       0,
	ToRate:         1,
	ExpectedResult: 0,
}, {
	Name:           "invalid (zero toRate)",
	Error:          "invalid rate: 0",
	FromAmount:     1,
	FromRate:       1,
	ToRate:         0,
	ExpectedResult: 0,
}, {
	Name:           "invalid (negative amount)",
	Error:          "invalid amount: must be greater than or equal to zero",
	FromAmount:     -1,
	FromRate:       1,
	ToRate:         1,
	ExpectedResult: 0,
}, {
	Name:           "zero input",
	FromAmount:     0,           // 0 input unit
	FromRate:       10000 * 1e8, // $10000 / input unit
	ToRate:         1e8,         // $1 / output unit
	ExpectedResult: 0,           // 0 output units
}, {
	Name:           "whole unit input",
	FromAmount:     1e8,         // 1 input unit
	FromRate:       10000 * 1e8, // $10000 / input unit
	ToRate:         1e8,         // $1 / output unit
	ExpectedResult: 10000 * 1e8, // 10000 output units
}, {
	Name:           "half unit input",
	FromAmount:     0.5e8,       // 1/2 input unit
	FromRate:       10000 * 1e8, // $10000 / input unit
	ToRate:         1e8,         // $1 / output unit
	ExpectedResult: 5000 * 1e8,  // 5000 output units
}, {
	Name:           "smallest unit input",
	FromAmount:     1,           // 1e-8 input unit
	FromRate:       10000 * 1e8, // $10000 / input unit
	ToRate:         1e8,         // $1 / output unit
	ExpectedResult: 10000,       // 10000e-8 output units
}, {
	Name:           "smallest unit input (truncated result)",
	FromAmount:     1,           // 1e-8 input unit
	FromRate:       1e8,         // $1 / input unit
	ToRate:         10000 * 1e8, // $10000 / output unit
	ExpectedResult: 0,           // 0 output units
}}

func TestConversions_Convert(t *testing.T) {
	for _, test := range conversionTests {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			observedResult, err := Convert(test.FromAmount, test.FromRate, test.ToRate)
			if len(test.Error) != 0 {
				assert.EqualError(err, test.Error)
				return
			}
			assert.Equal(test.ExpectedResult, observedResult)
			assert.Nil(err)
		})
	}
}
