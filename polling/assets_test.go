package polling_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/pegnet/pegnet/common"
	. "github.com/pegnet/pegnet/polling"
	"github.com/zpatrick/go-config"
)

// TestBasicPollingSources creates 8 polling sources. 5 have 1 asset, 3 have all.
// We then check to make sure the 5 sources are higher priority then the second 3.
// The second two have a priority order as well, and will be listed in the prioriy list
// for every asset.
func TestBasicPollingSources(t *testing.T) {
	end := 6
	// Create the unit test creator
	NewTestingDataSource = func(config *config.Config, source string) (IDataSource, error) {
		s := new(UnitTestDataSource)
		v, err := strconv.Atoi(string(source[8]))
		if err != nil {
			panic(err)
		}
		s.value = float64(v)
		s.assets = []string{common.AllAssets[v]}
		s.name = fmt.Sprintf("UnitTest%d", v)

		// Catch all
		if v >= end {
			s.assets = common.AllAssets[1:]
		}
		return s, nil
	}

	p := common.NewUnitTestConfigProvider()
	// The order of these retrieved is random since the settings are a map
	p.Data = `
[OracleDataSources]
  UnitTest1=1
  UnitTest2=2
  UnitTest3=3
  UnitTest4=4
  UnitTest5=5
  UnitTest6=6
  UnitTest7=7
  UnitTest8=8
`

	config := config.NewConfig([]config.Provider{p})

	s := NewDataSources(config)

	pa, err := s.PullAllPEGAssets()
	if err != nil {
		t.Error(err)
	}
	for i, asset := range common.AllAssets {
		v, ok := pa[asset]
		if !ok {
			t.Errorf("%s is missing", asset)
			continue
		}
		if i < end {
			if int(v.Value) != i {
				t.Errorf("Exp value %d, found %d for %s", i, int(v.Value), asset)
			}

			// Let's also check there is 4 sources
			if len(s.AssetSources[asset]) != 4 && asset != "PNT" {
				t.Errorf("exp %d sources for %s, found %d", 4, asset, len(s.AssetSources[asset]))
			}
		} else {
			if int(v.Value) != end {
				t.Errorf("Exp value %d, found %d for %s", end, int(v.Value), asset)
			}
			// Let's also check there is 3 sources
			if len(s.AssetSources[asset]) != 3 {
				t.Errorf("exp %d sources for %s, found %d", 3, asset, len(s.AssetSources[asset]))
			}
		}
	}
}

// UnitTestDataSource just reports the value for the supported assets
type UnitTestDataSource struct {
	value  float64
	assets []string
	name   string
}

func NewUnitTestDataSource(config *config.Config) (*UnitTestDataSource, error) {
	s := new(UnitTestDataSource)
	return s, nil
}

func (d *UnitTestDataSource) Name() string {
	return d.name
}

func (d *UnitTestDataSource) Url() string {
	return "https://unit.test/"
}

func (d *UnitTestDataSource) SupportedPegs() []string {
	return d.assets
}

func (d *UnitTestDataSource) FetchPegPrices() (peg PegAssets, err error) {
	peg = make(map[string]PegItem)

	timestamp := time.Now()
	for _, asset := range d.SupportedPegs() {
		peg[asset] = PegItem{Value: d.value, When: timestamp, WhenUnix: timestamp.Unix()}
	}

	return peg, nil
}

func (d *UnitTestDataSource) FetchPegPrice(peg string) (i PegItem, err error) {
	return FetchPegPrice(peg, d.FetchPegPrices)
}
