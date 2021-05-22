package polling_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/pegnet/pegnet/common"
	. "github.com/pegnet/pegnet/polling"
	"github.com/pegnet/pegnet/testutils"
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
		s, _ := testutils.NewUnitTestDataSource(config)
		v, err := strconv.Atoi(string(source[8]))
		if err != nil {
			panic(err)
		}
		s.Value = float64(v)
		s.Assets = []string{common.AllAssets[v]}
		s.SourceName = fmt.Sprintf("UnitTest%d", v)

		// Catch all
		if v >= end {
			s.Assets = common.AllAssets[1:]
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

	c := config.NewConfig([]config.Provider{common.NewDefaultConfigOptionsProvider(), common.NewDefaultConfigOptionsProvider(), p})

	s := NewDataSources(c)

	pa, err := s.PullAllPEGAssets(2)
	if err != nil {
		t.Error(err)
	}
	for i, asset := range common.AssetsV2 {
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
			if len(s.AssetSources[asset]) != 4 && asset != "PEG" {
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

	// Test the override mechanic
	t.Run("Test the override", func(t *testing.T) {
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

[OracleAssetDataSourcesPriority]
  # Specific coin overrides
  USD=UnitTest8
`
		c = config.NewConfig([]config.Provider{common.NewDefaultConfigOptionsProvider(), common.NewDefaultConfigOptionsProvider(), p})

		s = NewDataSources(c)
		pa, err := s.PullAllPEGAssets(1)
		if err != nil {
			t.Error(err)
		}

		if v, ok := pa["USD"]; !ok {
			t.Error("No usd")
		} else {
			if int(v.Value) != 8 {
				t.Error("Override failed")
			}
		}

		if len(s.AssetSources["USD"]) != 1 {
			t.Errorf("exp %d sources for %s, found %d", 1, "USD", len(s.AssetSources["USD"]))
		}

		if s.AssetSources["USD"][0] != "UnitTest8" {
			t.Errorf("Exp UnitTest8, got %s", s.AssetSources["USD"][0])
		}
	})
}

func TestDataSourceStaleness(t *testing.T) {
	ds := make([]IDataSource, 20)
	mapped := make(map[string]IDataSource)
	var names []string

	reference := time.Now()
	for i := 0; i < len(ds); i++ {
		s := new(testutils.UnitTestDataSource)
		s.Value = float64(i + 1)
		s.Assets = common.AllAssets
		s.SourceName = fmt.Sprintf("UnitTest%d", i)
		s.Timestamp = func() time.Time {
			return time.Now().Add(time.Duration(int(s.Value)*-1) * time.Minute)
		}

		ds[i] = s
		mapped[s.SourceName] = ds[i]
		names = append(names, s.SourceName)
	}

	// Set to 10m staleness
	d := NewDataSources(configWithStaleness("10m"))
	d.AssetSources["EUR"] = reverse(names)

	price, err := d.PullBestPrice("EUR", reference, mapped, 4)
	if err != nil {
		t.Error(err)
	}

	if price.Value != 10.0 {
		t.Error("Expected a value of 10, as the prior were stale")
	}

	// Make everything stale
	d = NewDataSources(configWithStaleness("0s"))
	d.AssetSources["EUR"] = reverse(names)

	price, err = d.PullBestPrice("EUR", reference, mapped, 4)
	if err != nil {
		t.Error(err)
	}

	if price.Value != 20.0 {
		t.Error("Expected a value of 20.0, as all are stale, but that is the highest priority")
	}

	// Make nothing stale
	d = NewDataSources(configWithStaleness("1h"))
	d.AssetSources["EUR"] = reverse(names)

	price, err = d.PullBestPrice("EUR", reference, mapped, 4)
	if err != nil {
		t.Error(err)
	}

	if price.Value != 20.0 {
		t.Error("Expected a value of 20.0, as nothing is stale")
	}

}

func configWithStaleness(d string) *config.Config {
	custom := common.NewUnitTestConfigProvider()
	custom.Data = fmt.Sprintf(`
[Oracle]
  StaleQuoteDuration = %s
`, d)
	config := config.NewConfig([]config.Provider{common.NewDefaultConfigOptionsProvider(),
		common.NewUnitTestConfigProvider(),
		custom})

	return config
}

func reverse(list []string) []string {
	rev := make([]string, len(list))
	for i, v := range list {
		rev[len(list)-i-1] = v
	}
	return rev
}

func TestTruncate(t *testing.T) {
	type Vector struct {
		Vector float64
		Exp4   float64
		Exp8   float64
	}
	vects := []Vector{
		{1, 1, 1},
		{1.123456789, 1.1234, 1.12345678},
		{1.12, 1.12, 1.12},
		{1.1267, 1.1267, 1.1267},
	}

	for _, v := range vects {
		if r := TruncateTo4(v.Vector); r != v.Exp4 {
			t.Errorf("t4 exp %f, got %f", v.Exp4, r)
		}
		if r := TruncateTo8(v.Vector); r != v.Exp8 {
			t.Errorf("t8 exp %f, got %f", v.Exp8, r)
		}
	}
}

// Names should be consistent in the config and returned by the datasource
func TestDataSourceNames(t *testing.T) {
	for name, d := range AllDataSources {
		if d.Name() != name {
			t.Error("Name user types does not match name returned")
		}
	}
}
