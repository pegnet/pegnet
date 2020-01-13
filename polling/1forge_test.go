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

var oneForgeResp = `[{"symbol":"EURUSD","bid":1.11373,"ask":1.11373,"price":1.11373,"timestamp":1578937601},{"symbol":"JPYUSD","bid":0.00909711,"ask":0.00909711,"price":0.00909711,"timestamp":1578937567},{"symbol":"GBPUSD","bid":1.29948,"ask":1.29948,"price":1.29948,"timestamp":1578937605},{"symbol":"CADUSD","bid":0.766349,"ask":0.766349,"price":0.766349,"timestamp":1578937598},{"symbol":"CHFUSD","bid":1.03052,"ask":1.03052,"price":1.03052,"timestamp":1578937593},{"symbol":"SGDUSD","bid":0.742561,"ask":0.742567,"price":0.742564,"timestamp":1578937579},{"symbol":"HKDUSD","bid":0.12867,"ask":0.128671,"price":0.12867,"timestamp":1578937548},{"symbol":"MXNUSD","bid":0.0531759,"ask":0.0531773,"price":0.0531766,"timestamp":1578937598},{"symbol":"XAUUSD","bid":1550.35,"ask":1550.35,"price":1550.35,"timestamp":1578937605},{"symbol":"XAGUSD","bid":18,"ask":18.003,"price":18.0015,"timestamp":1578937605},{"symbol":"AUDUSD","bid":0.69101,"ask":0.69102,"price":0.691015,"timestamp":1578937605},{"symbol":"NZDUSD","bid":0.66315,"ask":0.66318,"price":0.663165,"timestamp":1578937604},{"symbol":"SEKUSD","bid":0.105597,"ask":0.105603,"price":0.1056,"timestamp":1578937593},{"symbol":"NOKUSD","bid":0.112394,"ask":0.112402,"price":0.112398,"timestamp":1578937594},{"symbol":"RUBUSD","bid":0.0163236,"ask":0.0163244,"price":0.016324,"timestamp":1578937603},{"symbol":"ZARUSD","bid":0.0693118,"ask":0.0693222,"price":0.069317,"timestamp":1578937600},{"symbol":"TRYUSD","bid":0.170442,"ask":0.170463,"price":0.170453,"timestamp":1578937598}]
`
