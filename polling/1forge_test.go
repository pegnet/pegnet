// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package polling_test

import (
	"testing"
)

func TestFixed1ForgePeggedAssets(t *testing.T) {
	FixedDataSourceTest(t, "1Forge", []byte(oneForgeResp))
}

// Needs a key
//func TestActual1ForgePeggedAssets(t *testing.T) {
//	ActualDataSourceTest(t, "1Forge")
//}

var oneForgeResp = `[{"symbol":"EURUSD","bid":1.10911,"ask":1.10911,"price":1.10911,"timestamp":1565971649},{"symbol":"JPYUSD","bid":0.00940673,"ask":0.00940689,"price":0.00940681,"timestamp":1565971649},{"symbol":"GBPUSD","bid":1.21479,"ask":1.21481,"price":1.2148,"timestamp":1565971649},{"symbol":"CADUSD","bid":0.752808,"ask":0.752819,"price":0.752814,"timestamp":1565971649},{"symbol":"CHFUSD","bid":1.02032,"ask":1.02034,"price":1.02033,"timestamp":1565971649},{"symbol":"SGDUSD","bid":0.72198,"ask":0.722006,"price":0.721993,"timestamp":1565971649},{"symbol":"HKDUSD","bid":0.127504,"ask":0.127507,"price":0.127506,"timestamp":1565971649},{"symbol":"MXNUSD","bid":0.0511065,"ask":0.0511099,"price":0.0511082,"timestamp":1565971649},{"symbol":"XAUUSD","bid":1512.68,"ask":1512.79,"price":1512.735,"timestamp":1565971649},{"symbol":"XAGUSD","bid":17.142,"ask":17.147,"price":17.1445,"timestamp":1565971649},{"symbol":"BTCUSD","bid":10335.2,"ask":10349.36,"price":10342.28,"timestamp":1565971649},{"symbol":"BCHUSD","bid":308.341,"ask":311.448,"price":309.8945,"timestamp":1565971649},{"symbol":"LTCUSD","bid":74.81,"ask":75.5,"price":75.155,"timestamp":1565971649},{"symbol":"ETHUSD","bid":184.31,"ask":186.64,"price":185.475,"timestamp":1565971649},{"symbol":"DSHUSD","bid":93.5413,"ask":94.4832,"price":94.0122,"timestamp":1565971649}]`
