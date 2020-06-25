package staking

import (
	"context"
	"fmt"

	"github.com/pegnet/pegnet/spr"
	"github.com/zpatrick/go-config"
)

type ISPRMaker interface {
	NewSPR(ctx context.Context, dbht int32, config *config.Config, alert chan *spr.SPRs) (*spr.StakingPriceRecord, error)
}

// SPRMaker
// TODO: Should we change this at all?
type SPRMaker struct {
}

func NewSPRMaker() *SPRMaker {
	o := new(SPRMaker)
	return o
}

func (SPRMaker) NewSPR(ctx context.Context, dbht int32, config *config.Config, alert chan *spr.SPRs) (*spr.StakingPriceRecord, error) {
	return spr.NewSpr(ctx, dbht, config, alert)
}

type BlockingSPRMaker struct {
	n chan *spr.StakingPriceRecord
}

func NewBlockingSPRMaker() *BlockingSPRMaker {
	b := new(BlockingSPRMaker)
	b.n = make(chan *spr.StakingPriceRecord, 5)
	return b
}

// Drain everything from the channels
func (b *BlockingSPRMaker) Drain() {
ClearSPRChannel:
	for { // Drain anything remaining or return the height that matches
		select {
		case <-b.n:
		default:
			break ClearSPRChannel
		}
	}
}

func (b *BlockingSPRMaker) RecSPR(spr *spr.StakingPriceRecord) {
	b.n <- spr
}

func (b *BlockingSPRMaker) NewSPR(ctx context.Context, dbht int32, config *config.Config, alert chan *spr.SPRs) (*spr.StakingPriceRecord, error) {
	o := <-b.n
	if o == nil {
		return nil, fmt.Errorf("spr failed to be created")
	}
	if o.Dbht != dbht {
	DrainSPRLoop:
		for { // Drain anything remaining or return the height that matches
			select {
			case o := <-b.n:
				if o != nil && o.Dbht == dbht {
					return o, nil
				}
			default:
				break DrainSPRLoop
			}
		}
		return nil, fmt.Errorf("not the right height, exp %d found %d. %d in queue.", dbht, o.Dbht, len(b.n))
	}
	return o, nil
}
