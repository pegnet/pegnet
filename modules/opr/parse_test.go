package opr_test

import (
	"bytes"
	"math/rand"
	"reflect"
	"testing"

	"github.com/pegnet/pegnet/modules/opr"
	"github.com/pegnet/pegnet/modules/testutils"
)

func TestParse(t *testing.T) {
	t.Run("V1 Test Parse", func(t *testing.T) {
		testParse(t, 1)
	})
	t.Run("V2 Test Parse", func(t *testing.T) {
		testParse(t, 2)
	})
}

// testParse checks that an opr can be marshalled into the same content used to create it.
// This just checks the marshaling is consistent with itself
//
// For v1 the json could be in a different order, but GoLang's json marshal is consistent, so
// we can assume the resulting json will be the same.
func testParse(t *testing.T, version uint8) {
	for i := 0; i < 100; i++ {
		// Start with random valid content
		_, _, content := testutils.RandomOPR(version)
		o, err := opr.Parse(content)
		if err != nil {
			t.Error(err)
		}

		// We marshal the opr into data that should match the content
		data, err := o.Marshal()
		if err != nil {
			t.Error(err)
		} else {
			if bytes.Compare(data, content) != 0 {
				t.Errorf("the marshalled data is different than the original content")
			}
		}

		if uint8(o.GetType()) != version {
			t.Errorf("different version parsed than expected")
		}

		// Now parsing a second opr from the marshalled data should yield an identical copy of the opr
		o2, err := opr.Parse(data)
		if err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(o, o2) {
			t.Errorf("different oprs parsed from the same content")
		}
	}
}

// TestInvalidParse ensures a proper error is returned with invalid data
func TestInvalidParse(t *testing.T) {
	var err error

	// No data
	if _, err = opr.Parse(nil); err == nil {
		t.Errorf("expected a parse error, got none")
	}

	if _, err = opr.Parse([]byte{}); err == nil {
		t.Errorf("expected a parse error, got none")
	}

	if _, err = opr.Parse([]byte{0x00}); err == nil {
		t.Errorf("expected a parse error, got none")
	}

	// Random data
	for i := 0; i < 100; i++ {
		b := make([]byte, rand.Intn(500))
		rand.Read(b)
		if _, err = opr.Parse(b); err == nil {
			t.Errorf("expected a parse error, got none")
		}
	}

}
