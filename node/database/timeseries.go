package database

import (
	"fmt"
	"reflect"
	"time"

	"github.com/FactomProject/factom"

	"github.com/pegnet/pegnet/opr"

	"github.com/jinzhu/gorm"
)

var TimeSeriesArray []ITimeSeriesData

type ITimeSeriesData interface {
	Time() time.Time
	Height() int64
}

func FieldValue(item interface{}, field string) interface{} {
	value := reflect.ValueOf(item).FieldByName(field)
	return value.Interface()
}

// TimeSeries is the base of all time series
type TimeSeries struct {
	Timestamp   time.Time `gorm:"index"`
	BlockHeight int64     `gorm:"index;type:integer;primary_key"`
}

func (t TimeSeries) Time() time.Time {
	return t.Timestamp
}

func (t TimeSeries) Height() int64 {
	return t.BlockHeight
}

// DifficultyTimeSeries is the time series that has difficulty values for the graded set.
type DifficultyTimeSeries struct {
	TimeSeries
	// The most difficulty opr submitted in this block
	HighestDifficulty uint64 `sql:"type:int unsigned"`
	// LastGradedDifficulty is the last graded opr's difficulty (usually the 50th)
	// 	If we have less than 50, the index will be detailed here.
	LastGradedIndex      int
	LastGradedDifficulty uint64 `sql:"type:int unsigned"`
}

func HasHighBit(v uint64) bool {
	return v>>63 == 1
}

// BeforeCreate
// uint64's cannot have their highest bit set. We should always have our highest bit set, so we mask it our when we add
// and unset it when we pull it
func (d *DifficultyTimeSeries) BeforeCreate() (err error) {
	if !HasHighBit(d.HighestDifficulty) || !HasHighBit(d.LastGradedDifficulty) {
		// Highest bit is unset in one of the two... this should never happen
		return fmt.Errorf("highest bit is 0 in difficulty")
	}

	// Remove the top bit
	d.HighestDifficulty = d.HighestDifficulty & 0x7fffffffffffffff
	d.LastGradedDifficulty = d.LastGradedDifficulty & 0x7fffffffffffffff

	return
}

func (d *DifficultyTimeSeries) AfterFind() (err error) {
	// Add back the top bit
	d.HighestDifficulty = d.HighestDifficulty | 0x8000000000000000
	d.LastGradedDifficulty = d.LastGradedDifficulty | 0x8000000000000000
	return
}

// NetworkHashrateTimeSeries that estimates the network's hashrate based on
// the top50 oprs
type NetworkHashrateTimeSeries struct {
	TimeSeries
	// BasedOnBest is a network hashrate estimate based on the most difficult opr
	BasedOnBest float64

	// BasedOnLast is a network hashrate estimate based on the last graded opr
	BasedOnLast float64
}

// NumberOPRRecordsTimeSeries is the number of oprs submitted in a given block
type NumberOPRRecordsTimeSeries struct {
	TimeSeries
	NumberOfOPRs     int
	NumberGradedOPRs int
}

// AssetPricing
type AssetPricingTimeSeries struct {
	TimeSeries
	Asset string `gorm:"type:varchar;primary_key"`
	Price float64
}

type UniqueGradedCoinbasesTimeSeries struct {
	TimeSeries
	BiggestMiner           int // Coinbase with most records in top 50
	UniqueGradedCoinbases  int
	UniqueWinningCoinbases int
}

func AutoMigrateTimeSeries(db *gorm.DB) {
	db.AutoMigrate(&DifficultyTimeSeries{})
	db.AutoMigrate(&NetworkHashrateTimeSeries{})
	db.AutoMigrate(&NumberOPRRecordsTimeSeries{})
	db.AutoMigrate(&UniqueGradedCoinbasesTimeSeries{})

	db.AutoMigrate(&AssetPricingTimeSeries{})
}

func TimeSeriesFromOPRBlock(block *opr.OprBlock) (t TimeSeries, err error) {
	dblock, _, err := factom.GetDBlockByHeight(block.Dbht)
	if err != nil {
		return t, err
	}

	t.Timestamp = time.Unix(int64(dblock.Header.Timestamp)*60, 0)
	t.BlockHeight = block.Dbht
	return t, nil
}

func DifficultyTimeSeriesTimeSeriesFromOPRBlock(sorted []*opr.OraclePriceRecord, t TimeSeries) DifficultyTimeSeries {
	var n DifficultyTimeSeries
	n.TimeSeries = t

	// Grab the top
	n.HighestDifficulty = sorted[0].Difficulty
	i := lastWithHighBit(sorted)
	n.LastGradedIndex = i + 1
	n.LastGradedDifficulty = sorted[i].Difficulty
	return n
}

func NetworkHashrateTimeSeriesFromOPRBlock(sorted []*opr.OraclePriceRecord, t TimeSeries) NetworkHashrateTimeSeries {
	var n NetworkHashrateTimeSeries
	n.TimeSeries = t

	// Grab the top
	n.BasedOnBest = opr.EffectiveHashRate(sorted[0].Difficulty, 1)

	i := lastWithHighBit(sorted)
	n.BasedOnLast = opr.EffectiveHashRate(sorted[i].Difficulty, i+1)
	return n
}

// lastWithHighBit returns the first index with a high bit set
func lastWithHighBit(sorted []*opr.OraclePriceRecord) int {
	for i := len(sorted) - 1; i >= 0; i-- {
		if HasHighBit(sorted[i].Difficulty) {
			return i
		}
	}
	return -1
}
