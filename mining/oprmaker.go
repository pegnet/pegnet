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

func (b *BlockingOPRMaker) RecOPR(opr *opr.OraclePriceRecord) {
	b.n <- opr
}

func (b *BlockingOPRMaker) NewOPR(ctx context.Context, minerNumber int, dbht int32, config *config.Config, alert chan *opr.OPRs) (*opr.OraclePriceRecord, error) {
	o := <-b.n
	if o.Dbht != dbht {
	DrainOPRLoop:
		for { // Drain anything remaining
			select {
			case <-b.n:
			default:
				break DrainOPRLoop
			}
		}
		return nil, fmt.Errorf("not the right height, exp %d found %d", o.Dbht, dbht)
	}
	return o, nil
}
