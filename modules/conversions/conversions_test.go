package conversions_test

import (
	"fmt"
	"math"
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
	Name  string
	Error string
	XRate uint64
	YRate uint64
	X1    int64
	Y1    int64
	X2    int64
	Y2    int64
}

var conversionTestVectors = []ConversionTest{{
	Name:  "invalid (zero fromRate)",
	Error: "invalid rate: 0",
	XRate: 0,
	YRate: 1,
	X1:    1,
	Y1:    0,
}, {
	Name:  "invalid (zero toRate)",
	Error: "invalid rate: 0",
	XRate: 1,
	YRate: 0,
	X1:    1,
	Y1:    0,
}, {
	Name:  "invalid (negative amount)",
	Error: "invalid amount: must be greater than or equal to zero",
	XRate: 1,
	YRate: 1,
	X1:    -1,
	Y1:    0,
}, {
	Name:  "zero input",
	XRate: 10000 * 1e8, // $10000 / input unit
	YRate: 1e8,         // $1 / output unit
	X1:    0,           // 0 input unit
	Y1:    0,           // 0 output units
}, {
	Name:  "whole unit input",
	XRate: 10000 * 1e8, // $10000 / input unit
	YRate: 1e8,         // $1 / output unit
	X1:    1e8,         // 1 input unit
	Y1:    10000 * 1e8, // 10000 output units
	X2:    1e8,
	Y2:    10000 * 1e8,
}, {
	Name:  "half unit input",
	XRate: 10000 * 1e8, // $10000 / input unit
	YRate: 1e8,         // $1 / output unit
	X1:    0.5e8,       // 1/2 input unit
	Y1:    5000 * 1e8,  // 5000 output units
	X2:    0.5e8,
	Y2:    5000 * 1e8,
}, {
	Name:  "smallest unit input",
	XRate: 10000 * 1e8, // $10000 / input unit
	YRate: 1e8,         // $1 / output unit
	X1:    1,           // 1e-8 input unit
	Y1:    10000,       // 10000e-8 output units
	X2:    1,
	Y2:    10000,
}, {
	Name:  "smallest unit input (truncated result)",
	XRate: 1e8,         // $1 / input unit
	YRate: 10000 * 1e8, // $10000 / output unit
	X1:    1,           // 1e-8 input unit
	Y1:    0,           // 0 output units (due to truncation)
	X2:    0,           // 0 (due to previous truncation)
	Y2:    0,
}}

func TestConversions_Convert_Vectors(t *testing.T) {
	for _, test := range conversionTestVectors {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)

			observedY1, err := Convert(test.X1, test.XRate, test.YRate)
			if len(test.Error) != 0 {
				assert.EqualError(err, test.Error)
				return
			}
			require.NoError(t, err)
			assert.Equal(test.Y1, observedY1, "Unexpected result for conversion: X1 --> Y1")

			// Due to truncation in integer division, there is often error present in the
			// conversion from Y back to X. Thus, we check that it is within the expected
			// margin of error.
			observedX2, err := Convert(test.Y1, test.YRate, test.XRate)
			require.NoError(t, err)
			observedError := int64(math.Abs(float64(test.X1 - observedX2)))
			maxExpectedError := maxConversionError(test.X1, test.Y1)
			assert.True(observedError <= maxExpectedError, "Margin of error exceeded for conversion Y1 --> X2")

			observedY2, err := Convert(test.X2, test.XRate, test.YRate)
			require.NoError(t, err)
			assert.Equal(test.Y2, observedY2, "Unexpected result for conversion: X2 --> Y2")
			assert.Equal(observedY1, observedY2, "Y1 != Y2")
		})
	}
}

func TestConversions_Convert_Random(t *testing.T) {
	for i := 0; i < 100; i++ {
		expectedErrorString := ""
		xRate := rand.Uint64()
		yRate := rand.Uint64()
		if xRate == 0 || yRate == 0 {
			expectedErrorString = "invalid rate: 0"
		}
		x1 := rand.Int63n(1e17) // Arbitrary maximum of 1 billion units (1e9 * 1e8)
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
			observedError := int64(math.Abs(float64(x1 - x2)))
			maxExpectedError := maxConversionError(x1, y1)
			assert.True(observedError <= maxExpectedError, "Margin of error exceeded for conversion Y1 --> X2")

			y2, err := Convert(x2, xRate, yRate)
			require.NoError(t, err)
			observedError = int64(math.Abs(float64(y1 - y2)))
			assert.True(observedError <= maxExpectedError, "Margin of error exceeded for conversion X2 --> Y2")
		})
	}
}

func maxConversionError(x, y int64) int64 {
	if x == 0 || y == 0 {
		return 1
	}
	xBig := big.NewInt(x)
	yBig := big.NewInt(y)

	ratioXY := big.NewInt(0).Div(xBig, yBig)
	ratioYX := big.NewInt(0).Div(yBig, xBig)
	if ratioXY.Cmp(ratioYX) == -1 {
		return ratioYX.Int64() + 1
	}
	return ratioXY.Int64() + 1
}
