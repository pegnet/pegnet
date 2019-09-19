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

// TestClone ensures the clone command works gets us a good copy
func TestClone(t *testing.T) {
	t.Run("V1 Test Parse", func(t *testing.T) {
		testClone(t, 1)
	})
	t.Run("V2 Test Parse", func(t *testing.T) {
		testClone(t, 2)
	})
}

func testClone(t *testing.T, version uint8) {
	for i := 0; i < 100; i++ {
		_, _, content := testutils.RandomOPR(version)
		o1, _ := opr.Parse(content)
		o2 := o1.Clone()
		if !reflect.DeepEqual(o1, o2) {
			t.Errorf("clone is not a deep equal")
		}
	}
}

// Ensure we keep under 1EC
// This test is a bit rough since things like the nonce and id are variable length.
// But it is a good indicator if things go far over 1EC
func TestUnder1EC(t *testing.T) {
	t.Run("version 1", func(t *testing.T) {
		test1EC(t, 1)
	})

	t.Run("version 2", func(t *testing.T) {
		test1EC(t, 2)
	})
}

func test1EC(t *testing.T, version uint8) {
	_, extids, content := testutils.RandomOPR(version)
	o, _ := opr.Parse(content)
	testutils.PopulateRandomWinners(o)

	tl := 0
	for _, ext := range extids {
		tl += len(ext) + 2
	}

	data, err := o.Marshal()
	if err != nil {
		t.Error(err)
	}

	tl += len(data)

	if tl > 1024 {
		t.Errorf("opr entry is over 1kb, found %d bytes", tl)
	}
}
