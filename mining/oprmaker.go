package mining

import (
	"context"
	"fmt"

	"github.com/pegnet/pegnet/opr"
	"github.com/zpatrick/go-config"
)

type IOPRMaker interface {
	NewOPR(ctx context.Context, minerNumber int, dbht int32, config *config.Config, alert chan *opr.OPRs) (*opr.OraclePriceRecord, error)
}

// OPRMaker
// TODO: Should we change this at all?
type OPRMaker struct {
}

func NewOPRMaker() *OPRMaker {
	o := new(OPRMaker)
	return o
}

func (OPRMaker) NewOPR(ctx context.Context, minerNumber int, dbht int32, config *config.Config, alert chan *opr.OPRs) (*opr.OraclePriceRecord, error) {
	return opr.NewOpr(ctx, minerNumber, dbht, config, alert)
}

type BlockingOPRMaker struct {
	n chan *opr.OraclePriceRecord
}

func NewBlockingOPRMaker() *BlockingOPRMaker {
	b := new(BlockingOPRMaker)
	b.n = make(chan *opr.OraclePriceRecord, 5)
	return b
}

// Drain everything from the channels
func (b *BlockingOPRMaker) Drain() {
ClearOPRChannel:
	for { // Drain anything remaining or return the height that matches
		select {
		case <-b.n:
		default:
			break ClearOPRChannel
		}
	}
}

func (b *BlockingOPRMaker) RecOPR(opr *opr.OraclePriceRecord) {
	b.n <- opr
}

func (b *BlockingOPRMaker) NewOPR(ctx context.Context, minerNumber int, dbht int32, config *config.Config, alert chan *opr.OPRs) (*opr.OraclePriceRecord, error) {
	o := <-b.n
	if o == nil {
		return nil, fmt.Errorf("opr failed to be created")
	}
	if o.Dbht != dbht {
	DrainOPRLoop:
		for { // Drain anything remaining or return the height that matches
			select {
			case o := <-b.n:
				if o != nil && o.Dbht == dbht {
					return o, nil
				}
			default:
				break DrainOPRLoop
			}
		}
		return nil, fmt.Errorf("not the right height, exp %d found %d. %d in queue.", dbht, o.Dbht, len(b.n))
	}
	return o, nil
}
