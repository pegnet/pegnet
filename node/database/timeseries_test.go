package database_test

import (
	"math/rand"
	"testing"

	"github.com/pegnet/pegnet/node/database"
)

func TestDifficultyTimeSeries_Callbacks(t *testing.T) {
	for i := 0; i < 1000; i++ {
		x := rand.Uint64()
		// Ensure the top bit is set
		for {
			if x>>63 != 1 {
				x = x << 1
				x += 1
			} else {
				break
			}
		}

		d := database.DifficultyTimeSeries{
			HighestDifficulty:    x,
			LastGradedDifficulty: x,
		}

		err := d.BeforeCreate()
		if err != nil {
			t.Error(err)
		}

		err = d.AfterFind()
		if err != nil {
			t.Error(err)
		}

		if d.HighestDifficulty != x || d.LastGradedDifficulty != x {
			t.Errorf("uint64 changed with callbacks")
		}
	}
}

func TestFieldValue(t *testing.T) {
	d := database.DifficultyTimeSeries{
		HighestDifficulty:    1,
		LastGradedDifficulty: 1,
	}

	v := database.FieldValue(d, "HighestDifficulty")
	if av, ok := v.(uint64); !ok {
		t.Error("did not get right type")
	} else {
		if av != 1 {
			t.Error("did not get right value")
		}
	}
}
