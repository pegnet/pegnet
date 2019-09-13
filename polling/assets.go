package polling

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pegnet/pegnet/common"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

var dLog = log.WithField("id", "DataSources")

// AllDataSources is just a hard coded list of all the available assets. This list is copied in
// `NewDataSource`. These are the only two spots the list should be hard coded.
// The reason I have the `new(DataSource` is so I can get the name, url, and supported
// pegs from this map. It's useful in the cmdline to fetch the datasources dynamically.
var AllDataSources = map[string]IDataSource{
	"APILayer": new(APILayerDataSource),
	"CoinCap":  new(CoinCapDataSource),
	"FixedUSD": new(FixedUSDDataSource),
	// ExchangeRates is daily,  don't show people this
	//"ExchangeRates":     new(ExchangeRatesDataSource),
	"Kitco": new(KitcoDataSource),
	// OpenExchangeRates is hourly, but has commodities
	"OpenExchangeRates": new(OpenExchangeRatesDataSource),
	"CoinMarketCap":     new(CoinMarketCapDataSource),
	"FreeForexAPI":      new(FreeForexAPIDataSource),
	"1Forge":            new(OneForgeDataSource),
	"AlternativeMe":     new(AlternativeMeDataSource),
}

func AllDataSourcesList() []string {
	list := make([]string, len(AllDataSources))
	i := 0
	for k := range AllDataSources {
		list[i] = k
		i++
	}
	return list
}

// CorrectCasing converts the case insensitive to the correct casing for the exchange
// This is mainly to handle user input from the command line.
func CorrectCasing(in string) string {
	lowerIn := strings.ToLower(in)
	cased := AllDataSourcesList()
	for _, v := range cased {
		if lowerIn == strings.ToLower(v) {
			return v
		}
	}

	// For unit testing support, we need to include this.
	// Unit tests can create arbitrary 'DataSources'.
	// If you run this outside a unit test, the data source
	// will error out and fail to be made.
	r, _ := regexp.Compile("UnitTest[0-9]*")
	if r.Match([]byte(in)) {
		return "UnitTest"
	}

	return in
}

func NewDataSource(source string, config *config.Config) (IDataSource, error) {
	var ds IDataSource
	var err error

	// Make it case insensitive.
	switch CorrectCasing(source) {
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
	case "FreeForexAPI":
		ds, err = NewFreeForexAPIDataSource(config)
	case "1Forge":
		ds, err = NewOneForgeDataSourceDataSource(config)
	case "FixedUSD":
		ds, err = NewFixedUSDDataSource(config)
	case "AlternativeMe":
		ds, err = NewAlternativeMeDataSource(config)
	case "UnitTest": // This will fail outside a unit test
		ds, err = NewTestingDataSource(config, source)
	default:
		return nil, fmt.Errorf("%s is not a supported data source", source)
	}

	if err != nil {
		return nil, err
	}

	// Have 8min cache on all sources. Any more frequent queries will return the last
	// cached query. This is mainly for local testing on short blocks. We don't want to blow our
	// rate limits because we want to test on 30s blocks.
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

	config *config.Config
}

type DataSourceWithPriority struct {
	DataSource IDataSource
	Priority   int
}

// NewDataSources reads the config and sets everything up for polling
func NewDataSources(config *config.Config) *DataSources {
	d := new(DataSources)
	d.AssetSources = make(map[string][]string)
	d.DataSources = make(map[string]IDataSource)

	// All the config settings
	allSettings, err := config.Settings()
	common.CheckAndPanic(err)

	// We only want the OracleDataSources section. And they must be alpha numeric
	datasourceRegex, err := regexp.Compile(`OracleDataSources\.[a-zA-Z0-9]+`)
	common.CheckAndPanic(err)

	for setting := range allSettings {
		if datasourceRegex.Match([]byte(setting)) {
			// Get the priority. Priorities CANNOT be the same.
			p, err := config.Int(setting)
			common.CheckAndPanic(err)

			if p == -1 {
				continue // -1 priority means this source is disabled
			}

			source := strings.Split(setting, ".")
			if len(source) != 2 {
				panic(common.DetailError(fmt.Errorf("expect only 1 '.' in a setting. Found more : %s", setting)))
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

	// Validate that all priorities are only used once
	last := DataSourceWithPriority{Priority: -1}
	for _, v := range d.PriorityList {
		if v.Priority == last.Priority {
			dLog.Errorf("You may only use priorities once for your data sources in your config file. '%s' and '%s' both have '%d'",
				v.DataSource.Name(), last.DataSource.Name(), v.Priority)
			common.CheckAndPanic(fmt.Errorf("priority '%d' used more than once", v.Priority))
		}
		last = v
	}

	// Add the data sources
	// Yes I'm brute forcing it. Yes there is probably a better way. These lists are small
	// so it's not worth trying to get fancy. (3 nested for loops)
	for _, asset := range common.AllAssets { // For each asset we need
		for _, s := range d.PriorityList { // Append the data sources for that asset in priority order
			if common.FindIndexInStringArray(s.DataSource.SupportedPegs(), asset) != -1 {
				d.AssetSources[asset] = append(d.AssetSources[asset], s.DataSource.Name())
			}
		}
	}

	// Now we search for specific overrides. We will assume the user is a power user,
	// so we will not check if the data source has the asset or anything proper.
	// If they mess this up... well, they shouldn't be using this feature.
	// If someone wants to add validation here, go ahead :)
	for _, asset := range common.AllAssets {
		if order, err := config.String("OracleAssetDataSourcesPriority." + asset); err == nil && order != "" {
			d.AssetSources[asset] = strings.Split(order, ",")
		}
	}

	return d
}

// sortPriorityList sorts by priority
func (ds *DataSources) sortPriorityList() {
	sort.SliceStable(ds.PriorityList, func(i, j int) bool { return ds.PriorityList[i].Priority < ds.PriorityList[j].Priority })
}

// PriorityListString is for the cmd line, it will print all the data sources in their priority order.
func (ds *DataSources) PriorityListString() string {
	var str []string
	for _, v := range ds.PriorityList {
		str = append(str, fmt.Sprintf("%s:%d", v.DataSource.Name(), v.Priority))
	}

	if len(str) == 0 {
		return "You have no data sources configured"
	}
	return strings.Join(str, ", ")
}

// AssetPriorityString will print all the data sources for it in the priority order.
func (ds *DataSources) AssetPriorityString(asset string) string {
	str := ds.AssetSources[asset]
	if len(str) == 0 {
		return "NO DATASOURCE!"
	}
	return strings.Join(str, " -> ")
}

// PullAllPEGAssets will pull prices for every asset we are tracking.
// We pull assets from the sources in their priority order when possible.
// If an asset from priority 1 is missing, we resort to priority 2 ONLY for
// that missing asset.
// 	Params:
//		oprversion is passed in. Once we get past version 2, we can drop that from the
//					params and default to the version 2 behavior
// TODO: Currently we lazy eval prices, so we make the API call when we
//		first need a price from that source. These calls should be quick,
//		but it might be faster to eager eval all the data sources concurrently.
func (d *DataSources) PullAllPEGAssets(oprversion uint8) (pa PegAssets, err error) {
	assets := common.AllAssets // All the assets we are tracking.

	// Wrap all the data sources with a quick caching layer for
	// this loop. We only want to make 1 api call per source per Pull.
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
			// We found a price, so break out.
			if price.Value != 0 {
				break
			}
		}

		if err != nil { // This will only be the last err in the data source list.
			// No prices found for a peg, this pull failed
			return nil, fmt.Errorf("no price found for %s : %s", asset, err.Error())
		}

		// We round all prices to the same precision
		// Keep in mind if we didn't get a price (like no data sources), this will be a 0.
		// Validation of the price does not happen here. For example, PEG is an asset with no data source,
		// so it will be 0 here.
		// We WILL error out if all our data sources for a peg failed, and we listed data sources. That
		// is important to note.
		if oprversion == 1 {
			price.Value = TruncateTo4(price.Value)
		} else {
			price.Value = TruncateTo8(price.Value)
		}
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
