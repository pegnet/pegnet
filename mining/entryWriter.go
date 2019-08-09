package mining

import (
	"errors"
	"fmt"
	"sync"

	"github.com/FactomProject/factom"
	"github.com/cenkalti/backoff"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/opr"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

type IEntryWriter interface {
	PopulateECAddress() error
	NextBlockWriter() IEntryWriter
	AddMiner() chan<- *opr.NonceRanking
	SetOPR(opr *opr.OraclePriceRecord)
	CollectAndWrite(blocking bool)
	ECBalance() (int64, error)
}

// EntryWriter writes the best OPRs to factom once all the mining is done
type EntryWriter struct {
	Keep int
	// We need an opr template to make the entries
	oprTemplate *opr.OraclePriceRecord

	ec     *factom.ECAddress
	config *config.Config

	minerLists chan *opr.NonceRanking
	miners     int

	Next *EntryWriter

	EntryWritingFunction func(unique *opr.UniqueOPRData) error

	sync.Mutex
	sync.Once
}

func NewEntryWriter(config *config.Config, keep int) *EntryWriter {
	w := new(EntryWriter)
	w.Keep = keep
	w.minerLists = make(chan *opr.NonceRanking, keep)
	w.config = config
	w.EntryWritingFunction = w.writeMiningRecord

	return w
}

func (w *EntryWriter) ECBalance() (int64, error) {
	return factom.GetECBalance(w.ec.String())
}

// PopulateECAddress only needs to be called once
func (w *EntryWriter) PopulateECAddress() error {
	// Get the Entry Credit Address that we need to write our OPR records.
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

// AddMiner will add a miner to listen to for this block, and return the channel they
// should talk to us on.
func (w *EntryWriter) AddMiner() chan<- *opr.NonceRanking {
	w.Lock()
	defer w.Unlock()

	w.miners++
	return w.minerLists
}

// SetOPR is here because we need an opr to create the entry.
func (w *EntryWriter) SetOPR(opr *opr.OraclePriceRecord) {
	w.Lock()
	defer w.Unlock()
	if w.oprTemplate == nil {
		w.oprTemplate = opr.CloneEntryData()
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
	var aggregate []*opr.NonceRanking
GatherListLoop:
	for { // Collect all the miner submissions
		select {
		case list := <-w.minerLists:
			aggregate = append(aggregate, list)
			if len(aggregate) == w.miners {
				break GatherListLoop
			}
		}
	}

	// Merge miner submissions
	final := opr.MergeNonceRankings(w.Keep, aggregate...)
	nonces := final.GetNonces()
	for _, u := range nonces {
		err := w.EntryWritingFunction(u) // Write to blockchain
		if err != nil {
			log.WithError(err).Error("Failed to write mining record")
		}
	}

	dbht := int32(-1)
	if w.oprTemplate != nil {
		dbht = w.oprTemplate.Dbht
	}

	log.WithFields(log.Fields{
		"miner_count": w.miners,
		"height":      dbht,
		"exp_records": w.Keep,
		"records":     len(nonces),
	}).Info("OPR Block Mined")
}

// writeMiningRecord writes an opr and it's nonce to the blockchain
func (w *EntryWriter) writeMiningRecord(unique *opr.UniqueOPRData) error {
	if w.oprTemplate == nil {
		return fmt.Errorf("no opr template")
	}

	operation := func() error {
		var err1, err2 error
		entry, err := w.oprTemplate.CreateOPREntry(unique.Nonce, unique.Difficulty)
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
	w.miners--
	w.minerLists <- nil
}

// EntryForwarder is a wrapper for network based miners to rely on a coordinator to write entries
type EntryForwarder struct {
	*EntryWriter
	Next *EntryForwarder

	entryChannel chan *factom.Entry
}

func NewEntryForwarder(config *config.Config, keep int, entryChannel chan *factom.Entry) *EntryForwarder {
	n := new(EntryForwarder)
	n.EntryWriter = NewEntryWriter(config, keep)
	n.entryChannel = entryChannel
	n.EntryWritingFunction = n.forwardMiningRecord

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

func (w *EntryForwarder) forwardMiningRecord(unique *opr.UniqueOPRData) error {
	if w.oprTemplate == nil {
		return fmt.Errorf("no opr template")
	}

	entry, err := w.oprTemplate.CreateOPREntry(unique.Nonce, unique.Difficulty)
	if err != nil {
		return err
	}

	w.entryChannel <- entry
	return nil
}
