package testutils

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pegnet/pegnet/common"

	"github.com/pegnet/pegnet/polling"

	"github.com/zpatrick/go-config"
)

// AlwaysOnePolling returns 1 for all prices
func AlwaysOnePolling() *polling.DataSources {
	polling.NewTestingDataSource = func(config *config.Config, source string) (polling.IDataSource, error) {
		s, _ := NewUnitTestDataSource(config)
		v, err := strconv.Atoi(string(source[8]))
		if err != nil {
			panic(err)
		}
		s.Value = float64(1)
		s.Assets = common.AllAssets
		s.SourceName = fmt.Sprintf("UnitTest%d", v)

		return s, nil
	}

	p := common.NewUnitTestConfigProvider()
	p.Data = `
[OracleDataSources]
  UnitTest1=1
`
	c := config.NewConfig([]config.Provider{common.NewDefaultConfigOptionsProvider(), p})
	s := polling.NewDataSources(c)
	return s
}

// UnitTestDataSource just reports the Value for the supported Assets
type UnitTestDataSource struct {
	Value      float64
	Assets     []string
	SourceName string

	// How to timestamp price quotes
	Timestamp func() time.Time
}

func NewUnitTestDataSource(config *config.Config) (*UnitTestDataSource, error) {
	s := new(UnitTestDataSource)
	s.Timestamp = time.Now
	return s, nil
}

func (d *UnitTestDataSource) Name() string {
	return d.SourceName
}

func (d *UnitTestDataSource) Url() string {
	return "https://unit.test/"
}

func (d *UnitTestDataSource) SupportedPegs() []string {
	return d.Assets
}

func (d *UnitTestDataSource) FetchPegPrices() (peg polling.PegAssets, err error) {
	peg = make(map[string]polling.PegItem)

	timestamp := d.Timestamp()
	for _, asset := range d.SupportedPegs() {
		peg[asset] = polling.PegItem{Value: d.Value, When: timestamp, WhenUnix: timestamp.Unix()}
	}

	return peg, nil
}

func (d *UnitTestDataSource) FetchPegPrice(peg string) (i polling.PegItem, err error) {
	return polling.FetchPegPrice(peg, d.FetchPegPrices)
}
