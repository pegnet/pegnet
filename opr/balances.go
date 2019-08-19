// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package opr

//// Balances holds the non-zero balances for every address for every token in a
//// two dimensional map:
//// assetname => { RCD-hash => balance }
//var Balances map[string]map[[32]byte]int64
//
//func init() {
//	Balances = make(map[string]map[[32]byte]int64)
//	for _, asset := range common.AllAssets {
//		Balances["PNT"] = make(map[[32]byte]int64)
//		Balances["p"+asset] = make(map[[32]byte]int64)
//		Balances["t"+asset] = make(map[[32]byte]int64)
//	}
//}
//
//// ConvertAddress takes a human-readable address and extracts the prefix and RCD hash
//func ConvertAddress(address string) (prefix string, adr [32]byte, err error) {
//	prefix, adr2, err := common.ConvertPegNetAssetToRaw(address)
//	if err != nil {
//		return
//	}
//	copy(adr[:], adr2)
//	return
//}
//
//// AddToBalance adds the given value to the human-readable address
//// Note that add to balance takes a signed update, so this can be used to add to or
//// subtract from a balance.  An error is returned if the value would drive the balance
//// negative.  Or if the string doesn't represent a valid token
//func AddToBalance(address string, value int64) (err error) {
//	prefix, addressBytes, err := ConvertAddress(address)
//	if err != nil {
//		return errors.New("address not properly formed")
//	}
//	prev := Balances[prefix][addressBytes]
//	if prev+value < 0 {
//		return fmt.Errorf("result would be less than zero %d-%d", prev, -value)
//	}
//	Balances[prefix][addressBytes] = prev + value
//
//	//log.WithFields(log.Fields{
//	//	"address":       address,
//	//	"prefix":        prefix,
//	//	"address_bytes": hex.EncodeToString(addressBytes[:]),
//	//	"prev_balance":  prev,
//	//	"value":         value,
//	//	"new_balance":   prev + value,
//	//}).Debug("Add to balance")
//	return
//}
//
//// GetBalance returns the balance for a PegNet asset.
//// If the address is invalid, a -1 is returned
//func GetBalance(address string) (balance int64) {
//	prefix, adr, err := ConvertAddress(address)
//	if err != nil {
//		return -1
//	}
//	balance = Balances[prefix][adr]
//	return
//}
