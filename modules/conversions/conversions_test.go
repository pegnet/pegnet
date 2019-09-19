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
	Name:           "valid",
	FromAmount:     1e8,
	FromRate:       10250 * 1e8,
	ToRate:         1e8,
	ExpectedResult: 10250 * 1e8,
}, {
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
