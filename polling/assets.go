package polling

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pegnet/pegnet/common"
	log "github.com/sirupsen/logrus"
	config "github.com/zpatrick/go-config"
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
	"PegnetMarketCap":   new(PegnetMarketCapDataSource),
	"CoinGecko":         new(CoinGeckoDataSource),
	//"Factoshiio":        new(FactoshiioDataSource), // This will be deprecated
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
	case "PegnetMarketCap":
		ds, err = NewPegnetMarketCapDataSource()
	case "Factoshiio":
		ds, err = NewFactoshiioDataSource()
	case "CoinGecko":
		ds, err = NewCoinGeckoDataSource()
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

	// Some configuration variables read in from the config
	staleDuration time.Duration
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
	d.config = config

	// Load some specific config settings
	staleDuration, err := d.config.String(common.ConfigStaleDuration)
	if err != nil {
		common.CheckAndPanic(err)
	}

	duration, err := time.ParseDuration(staleDuration)
	if err != nil {
		common.CheckAndPanic(err)
	}
	d.staleDuration = duration

	// Load all the data-source config settings
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
	assets := common.AssetsV2 // All the assets we are tracking.
	if oprversion == 4 {
		assets = common.AssetsV4
	}
	if oprversion == 5 {
		assets = common.AssetsV5
	}
	start := time.Now()

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
		// For each asset we try and find the best price quote we can.
		price, err := d.PullBestPrice(asset, start, cacheWrap, oprversion)
		if err != nil { // This will only be the last err in the data source list.
			// No prices found for a peg, this pull failed
			return nil, fmt.Errorf("no price found for %s : %s", asset, err.Error())
		}

		if asset == "PEG" && oprversion == 3 && price.Value == 0 {
			return nil, fmt.Errorf("no price found for %s : %s", asset, fmt.Errorf("PEG has no value, check your datasources"))
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

// Get Trimmed Mean calculation
// https://www.investopedia.com/terms/t/trimmed_mean.asp
func TrimmedMean(data []PegItem, p int) float64 {
	sort.Slice(data, func(i, j int) bool {
		return data[i].Value < data[j].Value
	})

	length := len(data)
	if length <= 3 {
		return data[length/2].Value
	}

	sum := 0.0
	for i := p; i < length-p; i++ {
		sum = sum + data[i].Value
	}
	return sum / float64(length-2*p)
}

// PullBestPrice pulls the best asset price we can find for a given asset.
// Params:
//		asset		Asset to pull pricing data
//		reference	Time reference to determine 'staleness' from
//		sources		Map of datasources to pull the price quote from.
//		oprversion	OPR version
func (d *DataSources) PullBestPrice(asset string, reference time.Time, sources map[string]IDataSource, oprversion uint8) (pa PegItem, err error) {
	if sources == nil {
		// If our data sources passed in are nil, then we don't need to do cache wrapping.
		// We should always have sources passed in, aside from unit tests.
		sources = d.DataSources
	}

	// All the given data sources for the asset
	sourceList := d.AssetSources[asset]

	var prices []PegItem

	// Eval all datasources from the reference time
	for _, source := range sourceList {
		var price PegItem
		price, err = sources[source].FetchPegPrice(asset)
		if err != nil {
			continue
		}

		if price.Value != 0 {
			prices = append(prices, price)
		}
	}

	if oprversion == 5 {
		pricesClone := prices
		if len(pricesClone) > 0 {
			// We calculate the tolerance band based on trimmed mean.
			// If one datasource returns a price out of the defined tolerance band,
			// it will not allow that source’s price to be included in that block.
			tMean := TrimmedMean(pricesClone, 1)

			toleranceRate := 0.01
			if tMean >= 100000 {
				toleranceRate = 0.001
			}
			if tMean >= 1000 {
				toleranceRate = 0.01
			}
			if tMean < 1000 {
				toleranceRate = 0.1
			}
			toleranceBandHigh := tMean * (1 + toleranceRate)
			toleranceBandLow := tMean * (1 - toleranceRate)

			// We keep datasource priority order here.
			for i := 0; i < len(prices); i++ {
				currentPrice := prices[i]
				if currentPrice.Value >= toleranceBandLow && currentPrice.Value <= toleranceBandHigh {
					return currentPrice, nil
				}
			}
		}
	}

	// If we got here, that means that no price is passed by tolerance band.
	// We will take the highest priority quote given our data-source order.
	if len(prices) > 0 {
		pa = prices[0]
		return pa, nil
	}

	// If we have a detailed error on the last polling, we can just
	// return that
	if err != nil {
		return
	}

	// PEG is not configured on most miners.
	// TODO: When all miners start using FixedPEG as a datasource, we can
	// 		drop this conditional.
	if asset == "PEG" {
		return pa, nil
	}

	// If no prices exist, this could mean we had no data-sources configured for this
	// asset. We will return an error informing the user that this asset is not
	// configured. They should use a command like `pegnet datasources` to verify'
	// they have a proper config file.
	return pa, fmt.Errorf("'%s' doesn't seem to have any datasources configured. "+
		"please check your config file to ensure a datasource exists for this asset", asset)
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
