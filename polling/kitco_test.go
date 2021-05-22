// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package polling_test

import (
	"testing"
)

// TestKitcoPeggedAssets tests all the metals assets are found on kitco
func TestKitcoPeggedAssets(t *testing.T) {
	ActualDataSourceTest(t, "Kitco")
}

// The fixed is huge and a web scrape
