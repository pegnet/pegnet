package staking

import (
	"errors"
	"fmt"
	"github.com/pegnet/pegnet/spr"
	"sync"

	"github.com/FactomProject/factom"
	"github.com/cenkalti/backoff"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/opr"
	"github.com/zpatrick/go-config"
)

type IEntryWriter interface {
	PopulateECAddress() error
	NextBlockWriter() IEntryWriter
	SetSPR(spr *spr.StakingPriceRecord)
	CollectAndWrite(blocking bool)
	ECBalance() (int64, error)
}

// EntryWriter writes the best SPRs to factom once all the staking is done
type EntryWriter struct {
	Keep int
	// We need an spr template to make the entries
	sprTemplate *spr.StakingPriceRecord

	ec     *factom.ECAddress
	config *config.Config

	Next *EntryWriter

	EntryWritingFunction func(unique *opr.UniqueOPRData) error

	sync.Mutex
	sync.Once
}

func NewEntryWriter(config *config.Config, keep int) *EntryWriter {
	w := new(EntryWriter)
	w.Keep = keep
	w.config = config
	w.EntryWritingFunction = w.writeStakingRecord

	return w
}

func (w *EntryWriter) ECBalance() (int64, error) {
	return factom.GetECBalance(w.ec.String())
}

// PopulateECAddress only needs to be called once
func (w *EntryWriter) PopulateECAddress() error {
	// Get the Entry Credit Address that we need to write our SPR records.
	if ecadrStr, err := w.config.String("Miner.ECAddress"); err != nil {
		return err
	} else {
		ecAdr, err := factom.FetchECAddress(ecadrStr)
		if err != nil {
			return err
		}
		w.ec = ecAdr
	}
	return nil
}

// NextBlockWriter gets the next block writer to use for the miner.
//	Because all miners will share a block writer, we make this call idempotent
func (w *EntryWriter) NextBlockWriter() IEntryWriter {
	w.Lock()
	defer w.Unlock()
	if w.Next == nil {
		w.Next = NewEntryWriter(w.config, w.Keep)
		w.Next.ec = w.ec
	}
	return w.Next
}

// SetSPR is here because we need an spr to create the entry.
func (w *EntryWriter) SetSPR(spr *spr.StakingPriceRecord) {
	w.Lock()
	defer w.Unlock()
	if w.sprTemplate == nil {
		w.sprTemplate = spr.CloneEntryData()
	}
}

// CollectAndWrite will write the block when we collected all the miner data
//	The blocking is mainly for unit tests.
func (w *EntryWriter) CollectAndWrite(blocking bool) {
	w.Do(func() {
		if blocking {
			w.collectAndWrite()
		} else {
			go w.collectAndWrite()
		}
	})
}

// collectAndWrite is idempotent
func (w *EntryWriter) collectAndWrite() {
	//	var aggregate []*opr.NonceRanking
	//GatherListLoop:
	//	for { // Collect all the miner submissions
	//		select {
	//		case list := <-w.minerLists:
	//			aggregate = append(aggregate, list)
	//			if len(aggregate) == w.miners {
	//				break GatherListLoop
	//			}
	//		}
	//	}
	//
	//	// Merge miner submissions
	//	final := opr.MergeNonceRankings(w.Keep, aggregate...)
	//	nonces := final.GetNonces()
	//	for _, u := range nonces {
	//		err := w.EntryWritingFunction(u) // Write to blockchain
	//		if err != nil {
	//			log.WithError(err).Error("Failed to write mining record")
	//		}
	//	}
	//
	//	dbht := int32(-1)
	//	if w.sprTemplate != nil {
	//		dbht = w.sprTemplate.Dbht
	//	}
	//
	//	log.WithFields(log.Fields{
	//		"miner_count": w.miners,
	//		"height":      dbht,
	//		"exp_records": w.Keep,
	//		"records":     len(nonces),
	//	}).Info("OPR Block Mined")
}

// writeStakingRecord writes an spr and it's nonce to the blockchain
func (w *EntryWriter) writeStakingRecord(unique *opr.UniqueOPRData) error {
	if w.sprTemplate == nil {
		return fmt.Errorf("no spr template")
	}

	operation := func() error {
		var err1, err2 error
		entry, err := w.sprTemplate.CreateSPREntry(unique.Nonce, unique.Difficulty)
		if err != nil {
			return err
		}

		_, err1 = factom.CommitEntry(entry, w.ec)
		_, err2 = factom.RevealEntry(entry)
		if err1 == nil && err2 == nil {
			return nil
		}

		return errors.New("Unable to commit entry to factom")
	}

	err := backoff.Retry(operation, common.PegExponentialBackOff())
	if err != nil {
		// TODO: Handle error in retry
		return err
	}
	return nil
}

// Cancel will cancel a miner's write. If the miner was stopped, we should not expect his write
func (w *EntryWriter) Cancel() {
	//w.miners--
	//w.minerLists <- nil
}

// EntryForwarder is a wrapper for network based stakers to rely on a coordinator to write entries
type EntryForwarder struct {
	*EntryWriter
	Next *EntryForwarder

	entryChannel chan *factom.Entry
}

func NewEntryForwarder(config *config.Config, keep int, entryChannel chan *factom.Entry) *EntryForwarder {
	n := new(EntryForwarder)
	n.EntryWriter = NewEntryWriter(config, keep)
	n.entryChannel = entryChannel
	n.EntryWritingFunction = n.forwardStakingRecord

	return n

}

// ECBalance is always positive, the coordinator will stop us mining if he runs out
func (w *EntryForwarder) ECBalance() (int64, error) {
	return 1, nil
}

// NextBlockWriter gets the next block writer to use for the miner.
//	Because all miners will share a block writer, we make this call idempotent
func (w *EntryForwarder) NextBlockWriter() IEntryWriter {
	w.Lock()
	defer w.Unlock()
	if w.Next == nil {
		w.Next = NewEntryForwarder(w.config, w.Keep, w.entryChannel)
	}
	return w.Next
}

func (w *EntryForwarder) forwardStakingRecord(unique *opr.UniqueOPRData) error {
	if w.sprTemplate == nil {
		return fmt.Errorf("no spr template")
	}

	entry, err := w.sprTemplate.CreateSPREntry(unique.Nonce, unique.Difficulty)
	if err != nil {
		return err
	}

	w.entryChannel <- entry
	return nil
}
