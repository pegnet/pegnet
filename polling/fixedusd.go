package polling

import (
	"time"

	"github.com/zpatrick/go-config"
)

// FixedUSDDataSource is the datasource for USD.
// USD is always 1 USD = 1USD.
type FixedUSDDataSource struct {
	config *config.Config
}

func NewFixedUSDDataSource(config *config.Config) (*FixedUSDDataSource, error) {
	s := new(FixedUSDDataSource)
	s.config = config

	return s, nil
}

func (d *FixedUSDDataSource) Name() string {
	return "FixedUSD"
}

func (d *FixedUSDDataSource) Url() string {
	return "no-url"
}

func (d *FixedUSDDataSource) SupportedPegs() []string {
	return []string{"USD"}
}

func (d *FixedUSDDataSource) FetchPegPrices() (peg PegAssets, err error) {
	peg = make(map[string]PegItem)
	timestamp := time.Now()
	// The USD price is always 1
	peg["USD"] = PegItem{Value: 1, WhenUnix: timestamp.Unix(), When: timestamp}
	return
}

func (d *FixedUSDDataSource) FetchPegPrice(peg string) (i PegItem, err error) {
	return FetchPegPrice(peg, d.FetchPegPrices)
}
