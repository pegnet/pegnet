package polling

import (
	"fmt"
	"net/http"
	"time"
)

// NewHTTPClient is a variable so we can override it in unit tests.
// Some data sources might build clients off of this base one
var NewHTTPClient = func() *http.Client {
	return &http.Client{}
}

// IDataSource is the implementation all data sources need to adheer to.
type IDataSource interface {
	// Include some human friendly things.
	Name() string // Human friendly name
	Url() string  // Url to their website

	// FetchPegPrices is a rest based API call to the data source to fetch
	// the prices for the supported pegs.
	FetchPegPrices() (peg PegAssets, err error)

	// FetchPegPrice only fetches the price for a single peg
	FetchPegPrice(peg string) (PegItem, error)

	// SupportedPegs tells us what supported pegs the exchange supports.
	// All exchanges should have a list of pegs they support. This should
	// be defined up front.
	SupportedPegs() []string
}

// FetchPegPrice is because this implementation is the same for each exchange and GoLang's
// inheritance makes child structs referencing parent structs weird.
func FetchPegPrice(peg string, FetchPegPrices func() (peg PegAssets, err error)) (i PegItem, err error) {
	p, err := FetchPegPrices() // 99% of the time this fetches a cached value, and does not make a new api call
	if err != nil {
		return
	}

	item, ok := p[peg]
	if !ok {
		return i, fmt.Errorf("peg not found")
	}
	return item, nil
}

// TimedDataSourceCache will limit the number of calls by caching the results for X period of time
type TimedDataSourceCache struct {
	IDataSource

	LastCall      time.Time
	CacheDuration time.Duration
	Cache         PegAssets
}

func NewTimedDataSourceCache(s IDataSource, cacheLength time.Duration) IDataSource {
	d := new(TimedDataSourceCache)
	d.IDataSource = s
	d.CacheDuration = cacheLength

	return d
}

func (d *TimedDataSourceCache) FetchPegPrices() (peg PegAssets, err error) {
	if d.Cache != nil {
		// If the cache is set, check the time passed
		if time.Since(d.LastCall) < d.CacheDuration {
			return d.Cache, nil
		}
	}
	cache, err := d.IDataSource.FetchPegPrices()
	if err != nil {
		return nil, err
	}
	d.Cache = cache
	d.LastCall = time.Now()
	return cache, nil
}

func (d *TimedDataSourceCache) FetchPegPrice(peg string) (i PegItem, err error) {
	return FetchPegPrice(peg, d.FetchPegPrices)
}
