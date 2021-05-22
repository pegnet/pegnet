package conversions_test

import (
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	. "github.com/pegnet/pegnet/modules/conversions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ConversionTest represents a test that uses the given rates for X and Y to:
// 1) convert X1 to Y1
// 2) convert Y1 to X2
// 3) convert X2 to Y2
//
// Due to error in integer division, X1 will not always equal X2.
// Likewise, Y1 will not always equal Y2.
//
// The maximum error can be determined from the maximum ratio between X1 and Y1:
// `maxError := max(X1/Y1, Y1/X1) + 1`
// Where 1 is added to the result to account for truncation.
type ConversionTest struct {
	Name        string `json:"name"`
	ErrorString string `json:"error_string,omitempty"`
	XRate       uint64 `json:"x_rate"`
	YRate       uint64 `json:"y_rate"`
	X1          int64  `json:"x1"`
	Y1          int64  `json:"y1"`
	X2          int64  `json:"x2,omitempty"`
	Y2          int64  `json:"y2,omitempty"`
}

var conversionTestVectors = []string{
	`{"name": "invalid (zero fromRate)", "error_string": "invalid rate: 0", "x_rate": 0, "y_rate": 1, "x1": 1, "y1": 0}`,
	`{"name": "invalid (zero toRate)", "error_string": "invalid rate: 0", "x_rate": 1, "y_rate": 0, "x1": 1, "y1": 0}`,
	`{"name": "invalid (negative amount)", "error_string": "invalid amount: must be greater than or equal to zero", "x_rate": 1, "y_rate": 1, "x1": -1, "y1": 0}`,
	`{"name": "zero input", "x_rate": 1000000000000, "y_rate": 100000000, "x1": 0, "y1": 0}`,
	`{"name": "whole unit input", "x_rate": 1000000000000, "y_rate": 100000000, "x1": 100000000, "y1": 1000000000000, "x2": 100000000, "y2": 1000000000000}`,
	`{"name": "smallest unit input", "x_rate": 1000000000000, "y_rate": 100000000, "x1": 1, "y1": 10000, "x2": 1, "y2": 10000}`,
	`{"name": "smallest unit input (truncated result)", "x_rate": 100000000, "y_rate": 1000000000000, "x1": 1, "y1": 0, "x2": 0, "y2": 0}`,
	`{"name": "int64 overflow", "error_string": "integer overflow", "x_rate": 9223372036854775807, "y_rate": 1, "x1": 2, "y1": 0}`,
}

func TestConversions_Convert_Vectors(t *testing.T) {
	for _, testJSON := range conversionTestVectors {
		var test ConversionTest
		err := json.Unmarshal([]byte(testJSON), &test)
		require.NoError(t, err, "ConversionTest JSON Unmarshal raised an unexpected error")
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)

			observedY1, err := Convert(test.X1, test.XRate, test.YRate)
			if len(test.ErrorString) != 0 {
				assert.EqualError(err, test.ErrorString)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.Y1, observedY1, "Unexpected result for conversion: X1 --> Y1")

			// Due to truncation in integer division, there is often error present in the
			// conversion from Y back to X. Thus, we check that it is within the expected
			// margin of error.
			observedX2, err := Convert(test.Y1, test.YRate, test.XRate)
			require.NoError(t, err)
			observedError := abs(test.X1 - observedX2)
			maxExpectedError := maxConversionError(test.XRate, test.YRate)
			require.True(t, observedError <= maxExpectedError, "Margin of error exceeded for conversion Y1 --> X2")

			observedY2, err := Convert(test.X2, test.XRate, test.YRate)
			require.NoError(t, err)
			observedError = abs(test.Y1 - observedY2)
			assert.True(observedError <= maxExpectedError, "Margin of error exceeded for conversion X2 --> Y2")
		})
	}
}

func TestConversions_Convert_Random(t *testing.T) {
	for i := 0; i < 100; i++ {
		expectedErrorString := ""
		xRate := rand.Uint64() % (1e5 * 1e8 / 2) // Arbitrary maximum rate of $50k USD per unit
		yRate := rand.Uint64() % (1e5 * 1e8 / 2)
		if xRate == 0 || yRate == 0 {
			expectedErrorString = "invalid rate: 0"
		}
		x1 := rand.Int63n(1e9 * 1e8) // Arbitrary maximum of 1 billion units
		t.Run(fmt.Sprintf("Iteration %d", i), func(t *testing.T) {
			assert := assert.New(t)

			y1, err := Convert(x1, xRate, yRate)
			if len(expectedErrorString) != 0 {
				assert.EqualError(err, expectedErrorString)
				return
			}
			require.NoError(t, err)

			// Due to truncation in integer division, there is often error present in the
			// conversion from Y back to X. Thus, we check that it is within the expected
			// margin of error.
			x2, err := Convert(y1, yRate, xRate)
			require.NoError(t, err)
			observedError := abs(x1 - x2)
			maxExpectedError := maxConversionError(yRate, xRate)
			assert.True(observedError <= maxExpectedError, "Margin of error exceeded for Y1 --> X2: observedError=%d, maxError=%d", observedError, maxExpectedError)

			y2, err := Convert(x2, xRate, yRate)
			require.NoError(t, err)
			observedError = abs(y1 - y2)
			maxExpectedError = maxConversionError(xRate, yRate)
			assert.True(observedError <= maxExpectedError, "Margin of error exceeded for X2 --> Y2: observedError=%d, maxError=%d", observedError, maxExpectedError)
		})
	}
}

func maxConversionError(fromRate, toRate uint64) int64 {
	if fromRate == 0 || toRate == 0 || fromRate < toRate {
		return 1
	}
	fromRateBig := big.NewInt(0).SetUint64(fromRate)
	toRateBig := big.NewInt(0).SetUint64(toRate)

	ratioFromTo := big.NewInt(0).Div(fromRateBig, toRateBig)
	ratioToFrom := big.NewInt(0).Div(toRateBig, fromRateBig)
	if ratioFromTo.Cmp(ratioToFrom) == -1 {
		return ratioToFrom.Int64() + 1
	}
	return ratioFromTo.Int64() + 1
}

func abs(x int64) int64 {
	if x < 0 {
		return 0 - x
	}
	return x
}
