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

var oneForgeResp = `
[
    {
        "p": 1.17777,
        "a": 1.1778,
        "b": 1.17773,
        "s": "EUR/USD",
        "t": 1596229059490
    },
    {
        "p": 0.0094472,
        "a": 0.0094533,
        "b": 0.009441,
        "s": "JPY/USD",
        "t": 1596229032268
    },
    {
        "p": 1.3084,
        "a": 1.3085,
        "b": 1.3083,
        "s": "GBP/USD",
        "t": 1596229060197
    },
    {
        "p": 0.74581,
        "a": 0.74585,
        "b": 0.74576,
        "s": "CAD/USD",
        "t": 1596229061086
    },
    {
        "p": 1.0941784,
        "a": 1.0943309,
        "b": 1.0940258,
        "s": "CHF/USD",
        "t": 1596229060980
    },
    {
        "p": 0.7274818,
        "a": 0.7276749,
        "b": 0.7272886,
        "s": "SGD/USD",
        "t": 1596229061283
    },
    {
        "p": 0.1290231,
        "a": 0.12902726,
        "b": 0.12901894,
        "s": "HKD/USD",
        "t": 1596229051904
    },
    {
        "p": 0.04489,
        "a": 0.044895,
        "b": 0.044885,
        "s": "MXN/USD",
        "t": 1596229058680
    },
    {
        "p": 1975.34,
        "a": 1975.54,
        "b": 1975.14,
        "s": "XAU/USD",
        "t": 1596229059940
    },
    {
        "p": 24.396,
        "a": 24.408,
        "b": 24.383,
        "s": "XAG/USD",
        "t": 1596229053557
    },
    {
        "p": 0.7144,
        "a": 0.7146,
        "b": 0.7142,
        "s": "AUD/USD",
        "t": 1596229060640
    },
    {
        "p": 0.6631,
        "a": 0.66319,
        "b": 0.66301,
        "s": "NZD/USD",
        "t": 1596229058756
    },
    {
        "p": 0.113866,
        "a": 0.113885,
        "b": 0.113846,
        "s": "SEK/USD",
        "t": 1596229059983
    },
    {
        "p": 0.109855,
        "a": 0.110077,
        "b": 0.109633,
        "s": "NOK/USD",
        "t": 1596229059688
    },
    {
        "p": 0.01347,
        "a": 0.01347,
        "b": 0.01346,
        "s": "RUB/USD",
        "t": 1596213754236
    },
    {
        "p": 0.05857686,
        "a": 0.05862833,
        "b": 0.05852539,
        "s": "ZAR/USD",
        "t": 1596229057434
    },
    {
        "p": 0.14334898,
        "a": 0.14343004,
        "b": 0.14326791,
        "s": "TRY/USD",
        "t": 1596229053500
    }
]
`
