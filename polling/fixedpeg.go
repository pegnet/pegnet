package polling

import (
	"time"

	"github.com/zpatrick/go-config"
)

// FixedPEGDataSource is the datasource for PEG.
// Currently PEG is valued at 0 USD. This will be removed when PEG
// has value
type FixedPEGDataSource struct {
	config *config.Config
}

func NewFixedPEGDataSource(config *config.Config) (*FixedPEGDataSource, error) {
	s := new(FixedPEGDataSource)
	s.config = config

	return s, nil
}

func (d *FixedPEGDataSource) Name() string {
	return "FixedPEG"
}

func (d *FixedPEGDataSource) Url() string {
	return "no-url"
}

func (d *FixedPEGDataSource) SupportedPegs() []string {
	return []string{"PEG"}
}

func (d *FixedPEGDataSource) FetchPegPrices() (peg PegAssets, err error) {
	peg = make(map[string]PegItem)
	timestamp := time.Now()
	// The USD price is always 0
	peg["PEG"] = PegItem{Value: 0, WhenUnix: timestamp.Unix(), When: timestamp}
	return
}

func (d *FixedPEGDataSource) FetchPegPrice(peg string) (i PegItem, err error) {
	return FetchPegPrice(peg, d.FetchPegPrices)
}
