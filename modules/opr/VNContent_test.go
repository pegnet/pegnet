package opr_test

import (
	"encoding/hex"
	"testing"

	"github.com/pegnet/pegnet/modules/opr"
)

// TestOPRAccessors just ensures our accessors are correct...
// I know this is a dumb test, but cross package tests don't count towards coverage.
// To make them reasonable tests, we will use a vector test.
func TestOPRAccessors(t *testing.T) {
	// V1
	o, err := opr.Parse(V1Vector)
	if err != nil {
		t.Error(err)
	}

	if len(o.GetPreviousWinners()) != 10 {
		t.Errorf("exp 10 winners, found %d", len(o.GetPreviousWinners()))
	}

	// Some spot checks
	if o.GetPreviousWinners()[0] != "cc125ca8ae376453" && o.GetPreviousWinners()[9] != "49b391b4db98bfeb" {
		t.Error("Winners are incorrect")
	}

	if o.GetAddress() != "FA28wF5WSfXVJB5kzLTwsLiGRS4qgAo5iVg4k3fqVPQmNK7PoLhY" {
		t.Error("coinbase address is incorrect")
	}

	if o.GetID() != "oprv2testing" {
		t.Error("minerid is incorrect")
	}

	if o.GetHeight() != 96029 {
		t.Error("height is incorrect")
	}

	if o.GetType() != opr.V1 {
		t.Error("type is incorrect")
	}

	// The order should be fixed
	if o.GetOrderedAssetsFloat()[0].Value != float64(0.0) &&
		o.GetOrderedAssetsFloat()[1].Value != 1.0 &&
		o.GetOrderedAssetsFloat()[2].Value != 1.101 &&
		o.GetOrderedAssetsUint()[2].Value != uint64(1.101*1e8) {
		t.Error("price quotes incorrect")
	}

	// V2
	o, err = opr.Parse(V2Vector)
	if err != nil {
		t.Error(err)
	}

	if len(o.GetPreviousWinners()) != 25 {
		t.Errorf("exp 25 winners, found %d", len(o.GetPreviousWinners()))
	}

	// Some spot checks
	if o.GetPreviousWinners()[0] != "1ddcd640a224ab91" &&
		o.GetPreviousWinners()[9] != "bcf7c8c73143f07a" &&
		o.GetPreviousWinners()[24] != "4805056e178169f7" {
		t.Error("Winners are incorrect")
	}

	if o.GetAddress() != "FA28wF5WSfXVJB5kzLTwsLiGRS4qgAo5iVg4k3fqVPQmNK7PoLhY" {
		t.Error("coinbase address is incorrect")
	}

	if o.GetID() != "fccentral001" {
		t.Error("minerid is incorrect")
	}

	if o.GetHeight() != 96165 {
		t.Error("height is incorrect")
	}

	if o.GetType() != opr.V2 {
		t.Error("type is incorrect")
	}

	// The order should be fixed
	if o.GetOrderedAssetsUint()[0].Value != 0 &&
		o.GetOrderedAssetsUint()[1].Value != 1e8 &&
		o.GetOrderedAssetsUint()[2].Value != uint64(110643335) &&
		o.GetOrderedAssetsFloat()[2].Value != opr.Uint64ToFloat(110643335) {
		t.Error("price quotes incorrect")
	}
}

var (
	// Entryhash 9d62156f456fa8a9aec84ebb66d4ce0b4c1a5aa63ac3abb87e72086bcf65f30d on mainnet
	V1Vector = []byte(`{"coinbase":"FA28wF5WSfXVJB5kzLTwsLiGRS4qgAo5iVg4k3fqVPQmNK7PoLhY","dbht":96029,"winners":["cc125ca8ae376453","e33c2fad5bbff70e","1fb0bd5645dd1083","5ed7c49186928782","05397ae9ad8b5379","7c71f843ac5e0f27","b1493c6e89a631c6","6cd4861aa7eefe19","83b216cb7254f98d","49b391b4db98bfeb"],"minerid":"oprv2testing","assets":{"PNT":0,"USD":1,"EUR":1.101,"JPY":0.0092,"GBP":1.2328,"CAD":0.7578,"CHF":1.0072,"INR":0.0139,"SGD":0.7251,"CNY":0.1405,"HKD":0.1275,"KRW":0.0008,"BRL":0.2458,"PHP":0.0192,"MXN":0.0512,"XAU":1497.0059,"XAG":18.1146,"XPD":1576.0689,"XPT":944.5011,"XBT":10093.9982,"ETH":176.918,"LTC":68.7392,"RVN":0.0302,"XBC":296.7023,"FCT":3.0989,"BNB":20.5529,"XLM":0.0581,"ADA":0.0445,"XMR":73.3111,"DASH":82.035,"ZEC":44.2895,"DCR":22.6989}}`)

	// Entryhash af03e85c4ca545f49ea15cc3e6cc94948033aad8d7c3c20cb057b7788e4091f8 on testnet
	V2Vector, _ = hex.DecodeString("0a344641323877463557536658564a42356b7a4c5477734c69475253347167416f35695667346b3366715650516d4e4b37506f4c6859120c666363656e7472616c30303118a5ef0522081ddcd640a224ab912208790e75cc2149d0cd2208c16e6fc535c7e773220869873f7e66c478ac2208a3d8bbc64bff03802208a068b60c77acaf902208edd1ad0137696c932208d764e6f5cb8c5afc2208bcf7c8c73143f07a22080c3f39055453824722084af344446a49e71c2208f4e12e587853150d2208d80d737085d803db22085b68fb2d0e03cb962208755c8db74d8563182208c5e0746b723e5c952208e49f3ba60d0665dc22082abc7b91786d34952208147aa04e113dfe9c2208883c651dd29d2ca02208f098ffd9954053c12208a4c2957c87fe70dd2208b3c87161bf02e18222086a1a17046a64356a22084805056e178169f72a7f0080c2d72f8791e1349db738ddcaea3a9c848a24c9d88f30b1fe55c180d422ac93de06c6888c06f89405addedf0b95e175df87ba02a98beac1ae04d296dddd06e0b39ee18c1e93ebac9e43abb791e219a8a8bb01b696c9b870fe95b89e01b3c08ae10790cbe802cfd49902c1b78dd51b818596a620f5a6ffce10d29d9cd308")
)
