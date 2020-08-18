package opr_test

import (
	"bytes"
	"math/rand"
	"reflect"
	"testing"

	lxr "github.com/pegnet/LXRHash"

	"github.com/pegnet/pegnet/modules/opr"
	"github.com/pegnet/pegnet/modules/testutils"
)

func init() {
	testutils.SetTestLXR(lxr.Init(lxr.Seed, 8, lxr.HashSize, lxr.Passes))
}

func TestParse(t *testing.T) {
	t.Run("V1 Test Parse", func(t *testing.T) {
		testParse(t, 1)
	})
	t.Run("V2 Test Parse", func(t *testing.T) {
		testParse(t, 2)
	})
	t.Run("V3 Test Parse", func(t *testing.T) {
		testParse(t, 3)
	})
	t.Run("V4 Test Parse", func(t *testing.T) {
		testParse(t, 4)
	})
	t.Run("V5 Test Parse", func(t *testing.T) {
		testParse(t, 5)
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
			// V3 and V4 both use the same content type as V2
			if (version == 3 || version == 4 || version == 5) && o.GetType() != opr.V2 {
				t.Errorf("different version parsed than expected")
			}
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
	t.Run("V1 Test Clone", func(t *testing.T) {
		testClone(t, 1)
	})
	t.Run("V2 Test Clone", func(t *testing.T) {
		testClone(t, 2)
	})
	t.Run("V3 Test Clone", func(t *testing.T) {
		testClone(t, 3)
	})
	t.Run("V4 Test Clone", func(t *testing.T) {
		testClone(t, 4)
	})
	t.Run("V5 Test Clone", func(t *testing.T) {
		testClone(t, 5)
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

	t.Run("version 3", func(t *testing.T) {
		test1EC(t, 3)
	})

	t.Run("version 4", func(t *testing.T) {
		test1EC(t, 4)
	})

	t.Run("version 5", func(t *testing.T) {
		test1EC(t, 5)
	})
}

func test1EC(t *testing.T, version uint8) {
	_, extids, content := testutils.RandomOPRWithRandomWinners(version, rand.Int31())
	o, _ := opr.Parse(content)

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

// TestStrangeVector
// This json vector does not throw a decode error on the protobuf v2.
func TestStrangeVector(t *testing.T) {
	vector := []byte(`{"coinbase":"FA2Q2qfexXN9xPnTo8QBCsbcSUCEieUtcd2PLMvNytHfov8rUyaL","dbht":1390713771,"winners":["","","","","","","","","",""],"minerid":"e37cdb9d640c077d","assets":{"PNT":0.8497,"USD":0.0806,"EUR":0.0596,"JPY":0.9798,"GBP":0.5676,"CAD":0.922,"CHF":0.8385,"INR":0.2486,"SGD":0.0765,"CNY":0.4301,"HKD":0.7849,"KRW":0.7163,"BRL":0.6036,"PHP":0.4904,"MXN":0.5164,"XAU":0.7571,"XAG":0.9346,"XPD":0.2451,"XPT":0.6606,"XBT":0.6204,"ETH":0.0908,"LTC":0.0124,"RVN":0.47,"XBC":0.6753,"FCT":0.4592,"BNB":0.1782,"XLM":0.9339,"ADA":0.0861,"XMR":0.581,"DASH":0.6887,"ZEC":0.216,"DCR":0.751}}`)

	o, err := opr.Parse(vector)
	if err != nil {
		t.Error(err)
	}

	if o.GetType() != opr.V1 {
		t.Error("exp v1")
	}

	// Here is the strange thing
	// This all works. The ParseV2 goes through, and
	// creates some opr. It's missing values for many fields
	// but the lack of error means the general parse should do V1 then V2
	o2, err := opr.ParseV2Content(vector)
	if err != nil {
		t.Error(err)
	}

	if o2.GetType() != opr.V2 {
		t.Error("exp v2")
	}
}
