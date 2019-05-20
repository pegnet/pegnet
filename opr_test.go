package oprecord_test

import (
	"fmt"
	oprecord "github.com/pegnet/OracleRecord"
	"testing"
)

func TestOPRmarshal(t *testing.T) {
	opr := new(oprecord.OraclePriceRecord)
	_ = opr

	opr.SetCoinbasePNTAddress([]byte("tPNT2VSeR9ga586m3q85JWruniRjnpjknLtyaj1eJ4X4gnEzbe76b103"))
	data, _ := opr.MarshalBinary()
	opr2 := new(oprecord.OraclePriceRecord)
	opr2.UnmarshalBinary(data)
	fmt.Println(opr.String())
	fmt.Println(opr2.String())
}
