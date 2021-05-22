package polling

import (
	"time"

	"github.com/pegnet/pegnet/common"
)

// IsMarketOpen is used to determine is a stale quote is due to a closed market.
// If the market is closed, we will allow stale quotes without having to fallback to
// a backup datasource. This 'MarketOpen' does not have to be perfect, as if they market
// is closed, but say it is open, then the polling will return the most recent quote it can find.
// This is not bad behavior.
func IsMarketOpen(asset string, reference time.Time) bool {
	if common.AssetListContains(common.CurrencyAssets, asset) {
		return ForexOpen(reference)
	}
	if common.AssetListContains(common.CommodityAssets, asset) {
		return CommodityOpen(reference)
	}
	return true
}

// ForexOpen returns if the forex markets are open
//	This is a rough estimate, which we use to allow stale quote data.
// https://www.forex.com/en-us/support/trading-hours/
// Opens:
//		5:00 pm ET Sunday -> 21:00 UTC Sunday
//	Closes:
//		5:00 pm ET Friday -> 21:00 UTC Friday
func ForexOpen(reference time.Time) bool {
	utc := reference.UTC()
	switch utc.Weekday() {
	case time.Saturday:
		return false
	case time.Friday:
		return reference.Hour() < 21
	case time.Sunday:
		return reference.Hour() >= 21
	}
	return true // Open Mon-Thurs for sure
}

// CommodityOpen returns if the commodity markets are open
//	This is a rough estimate, which we use to allow stale quote data.
//	There is a 1hr window where it is closed during the day, however I do
// 	not see that on chain, and a fallback to the next datasource is not a bad
//	result of letting the stale quote fallback.
// https://www.forex.com/en-us/support/trading-hours/
// Opens:
//		6:00 pm ET Sunday -> 22:00 UTC Sunday
//	Closes:
//		5:00 pm ET Friday -> 21:00 UTC Friday
func CommodityOpen(reference time.Time) bool {
	utc := reference.UTC()
	switch utc.Weekday() {
	case time.Saturday:
		return false
	case time.Friday:
		// Closes at 21:00
		return reference.Hour() < 21
	case time.Sunday:
		// Opens at 22:00
		return reference.Hour() >= 22
	}
	return true // Open Mon-Thurs for sure
}
