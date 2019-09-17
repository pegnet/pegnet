package conversions_test

import (
	"testing"

	. "github.com/pegnet/pegnet/modules/conversions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testsMinimumInputNeeded = []struct {
	Name            string
	Error           string
	Rates           OraclePriceRecordAssetList
	InputType       string
	OutputType      string
	RequestedOutput int64
	ExpectedResult  int64
}{{
	Name:       "valid",
	InputType:  "XBT",
	OutputType: "USD",
	Rates: OraclePriceRecordAssetList{
		"XBT": 10250 * 1e8,
		"USD": 1e8,
	},
	RequestedOutput: 50 * 1e8,
	ExpectedResult:  487804, // ($10250 / $50) * 1e8
}}

func TestConversions_MinimumInputNeeded(t *testing.T) {
	for _, test := range testsMinimumInputNeeded {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			observedResult, err := test.Rates.MinimumInputNeeded(test.InputType, test.OutputType, test.RequestedOutput)
			if len(test.Error) != 0 {
				if assert.NotNil(err) {
					assert.Contains(err.Error(), test.Error)
				}
				return
			}
			assert.Equal(test.ExpectedResult, observedResult)
			require.NoError(t, err)
		})
	}
}

var testsMaximumOutputPossible = []struct {
	Name           string
	Error          string
	Rates          OraclePriceRecordAssetList
	InputType      string
	OutputType     string
	AvailableInput int64
	ExpectedResult int64
}{{
	Name:       "valid",
	InputType:  "XBT",
	OutputType: "USD",
	Rates: OraclePriceRecordAssetList{
		"XBT": 10250 * 1e8,
		"USD": 1e8,
	},
	AvailableInput: 10 * 1e8,
	ExpectedResult: 10250 * 1e8 * 10,
}}

func TestConversions_MaximumOutputPossible(t *testing.T) {
	for _, test := range testsMaximumOutputPossible {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			observedResult, err := test.Rates.MaximumOutputPossible(test.InputType, test.AvailableInput, test.OutputType)
			if len(test.Error) != 0 {
				if assert.NotNil(err) {
					assert.Contains(err.Error(), test.Error)
				}
				return
			}
			assert.Equal(test.ExpectedResult, observedResult)
			require.NoError(t, err)
		})
	}
}
