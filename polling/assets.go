package polling

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pegnet/pegnet/common"
	"github.com/zpatrick/go-config"
)

func NewDataSource(source string, config *config.Config) (IDataSource, error) {
	var ds IDataSource
	var err error
	switch source {
	case "APILayer":
		ds, err = NewAPILayerDataSource(config)
	case "CoinCap":
		ds, err = NewCoinCapDataSource(config)
	case "ExchangeRates":
		ds, err = NewExchangeRatesDataSource(config)
	case "Kitco":
		ds, err = NewKitcoDataSource(config)
	case "OpenExchangeRates":
		ds, err = NewOpenExchangeRatesDataSource(config)
	case "CoinMarketCap":
		ds, err = NewCoinMarketCapDataSource(config)
	default:
		return nil, fmt.Errorf("%s is not a supported data source", source)
	}

	if err != nil {
		return nil, err
	}

	// Have 8min cache on all sources. Any more frequent queries will return the last
	// cached query
	return NewTimedDataSourceCache(ds, time.Minute*8), nil
}

// DataSources will initialize all data sources and handle pulling of all the assets.
type DataSources struct {
	// AssetSources are listed in priority order.
	// If the asset is missing, we will walk through the data sources to find the
	// price of an asset
	//	Key -> CurrencyISO
	//	Value -> List of data sources by name
	AssetSources map[string][]string

	// DataSources is all the data sources we support
	//	Key -> Data source name
	//	Value -> Data source struct
	DataSources map[string]IDataSource

	// The list of data sources by priority.
	PriorityList []DataSourceWithPriority

	maxPriority int // Each peg has a list of prioritized asset sources

	config *config.Config
}

type DataSourceWithPriority struct {
	DataSource IDataSource
	Priority   int
}

func NewDataSources(config *config.Config) *DataSources {
	d := new(DataSources)
	d.AssetSources = make(map[string][]string)
	d.DataSources = make(map[string]IDataSource)

	// Create data sources
	allSettings, err := config.Settings()
	common.CheckAndPanic(err)

	datasourceRegex, err := regexp.Compile(`OracleDataSources\.[a-zA-Z0-9]+`)
	common.CheckAndPanic(err)

	for setting, _ := range allSettings {
		if datasourceRegex.Match([]byte(setting)) {
			// Get the priority. Priorities can be the same, then we'll sort
			// alphabetically to keep the results deterministic
			p, err := config.Int(setting)
			common.CheckAndPanic(err)

			if p == -1 {
				continue // This source is disabled
			}

			source := strings.Split(setting, ".")
			if len(source) != 2 {
				panic(common.DetailError(fmt.Errorf("expect only 1 '.' in a setting. Found %s", setting)))
			}
			s, err := NewDataSource(source[1], config)
			common.CheckAndPanic(err)

			// Add to our lists
			d.PriorityList = append(d.PriorityList, DataSourceWithPriority{DataSource: s, Priority: p})
			d.DataSources[s.Name()] = s
		}
	}

	// Ensure it is sorted
	d.sortPriorityList()

	// Add the data sources
	// Yes I'm brute forcing it. Yes there is probably a better way. These lists are small
	for _, asset := range common.AllAssets { // For each asset we need
		for _, s := range d.PriorityList { // Append the data sources for that asset in priority order
			if common.StringArrayContains(s.DataSource.SupportedPegs(), asset) != -1 {
				d.AssetSources[asset] = append(d.AssetSources[asset], s.DataSource.Name())
			}
		}
	}

	return d
}

// sortPriorityList sorts by name, then by priority
func (ds *DataSources) sortPriorityList() {
	sort.SliceStable(ds.PriorityList, func(i, j int) bool {
		return ds.PriorityList[i].DataSource.Name() < ds.PriorityList[j].DataSource.Name()
	})
	sort.SliceStable(ds.PriorityList, func(i, j int) bool { return ds.PriorityList[i].Priority < ds.PriorityList[j].Priority })
}

// PullAllPEGAssets will pull prices for every asset we are tracking.
// We pull assets from the sources in their priority order when possible.
// If an asset from priority 1 is missing, we resort to priority 2 ONLY for
// that missing asset.
// TODO: Currently we lazy eval prices, so we make the API call when we
//		first need a price from that source. These calls should be quick,
//		but it might be faster to eager eval all the data sources concurrently.
func (d *DataSources) PullAllPEGAssets() (pa PegAssets, err error) {
	assets := common.AllAssets // All the assets we are tracking.

	// Wrap all the data sources with a quick caching layer for
	// this loop
	cacheWrap := make(map[string]IDataSource)

	for _, source := range d.DataSources {
		cacheWrap[source.Name()] = NewCachedDataSource(source)
	}

	// TODO: You would eager eval here, and block the for loop on the first data source that has
	// 		not completed. Or block the for loop until they all completed... but I'd prefer the former.

	pa = make(PegAssets)
	for _, asset := range assets {
		var price PegItem
		// For each asset we try the data source in the list.
		// If we find a price, we can exit early, as we only need 1 asset price
		// per peg.
		for _, sourceName := range d.AssetSources[asset] {
			price, err = cacheWrap[sourceName].FetchPegPrice(asset)
			if err != nil {
				continue // Try the next source
			}
		}

		if err != nil { // This will only be the last err in the data source list.
			// No prices found for a peg, this pull failed
			return nil, fmt.Errorf("no price found for %s : %s", asset, err.Error())
		}

		// We round all prices to the same precision
		price.Value = Round(price.Value)
		pa[asset] = price
	}

	return pa, nil
}

// OneTimeUseCache will cache the data source response so we can query for
// prices individually without making a new api call. This cache is only to be used
// for 1 moment, rather than being used as a cache across multiple actions/rountines.
type OneTimeUseCache struct {
	IDataSource
	Cache PegAssets
}

func NewCachedDataSource(d IDataSource) *OneTimeUseCache {
	c := new(OneTimeUseCache)
	c.IDataSource = d

	return c
}

func (d *OneTimeUseCache) FetchPegPrices() (peg PegAssets, err error) {
	if d.Cache != nil {
		return d.Cache, nil
	}
	cache, err := d.IDataSource.FetchPegPrices()
	if err != nil {
		return nil, err
	}
	d.Cache = cache
	return cache, nil
}

func (d *OneTimeUseCache) FetchPegPrice(peg string) (i PegItem, err error) {
	return FetchPegPrice(peg, d.FetchPegPrices)
}
