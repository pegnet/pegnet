package staking

import (
	"context"
	"github.com/pegnet/pegnet/spr"
	"github.com/zpatrick/go-config"
)

type ISPRMaker interface {
	NewSPR(ctx context.Context, dbht int32, config *config.Config) (*spr.StakingPriceRecord, error)
}

// SPRMaker
// TODO: Should we change this at all?
type SPRMaker struct {
}

func NewSPRMaker() *SPRMaker {
	o := new(SPRMaker)
	return o
}

func (SPRMaker) NewSPR(ctx context.Context, dbht int32, config *config.Config) (*spr.StakingPriceRecord, error) {
	return spr.NewSpr(ctx, dbht, config)
}
