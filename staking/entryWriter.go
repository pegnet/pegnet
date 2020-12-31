package staking

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/pegnet/pegnet/spr"
	log "github.com/sirupsen/logrus"

	"github.com/FactomProject/factom"
	"github.com/cenkalti/backoff"
	"github.com/pegnet/pegnet/common"
	"github.com/zpatrick/go-config"
)

type IEntryWriter interface {
	PopulateECAddress() error
	NextBlockWriter() IEntryWriter
	SetSPR(spr *spr.StakingPriceRecord)
	CollectAndWrite(blocking bool)
	ECBalance() (int64, error)
}

// EntryWriter writes the SPRs to factom once the staking is done
type EntryWriter struct {
	// We need an spr template to make the entries
	sprTemplate *spr.StakingPriceRecord

	ecUser string
	ec     *factom.ECAddress
	config *config.Config

	Next *EntryWriter

	EntryWritingFunction func() error

	sync.Mutex
	sync.Once
}

func NewEntryWriter(config *config.Config) *EntryWriter {
	w := new(EntryWriter)
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
	if ecadrStr, err := w.config.String("Staker.ECAddress"); err != nil {
		return err
	} else {
		ecAdr, err := factom.FetchECAddress(ecadrStr)
		if err != nil {
			return err
		}
		w.ec = ecAdr
		w.ecUser = ecadrStr
	}
	return nil
}

// NextBlockWriter gets the next block writer to use for the staker.
func (w *EntryWriter) NextBlockWriter() IEntryWriter {
	w.Lock()
	defer w.Unlock()
	if w.Next == nil {
		w.Next = NewEntryWriter(w.config)
		w.Next.ecUser = w.ecUser
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

// CollectAndWrite will write the block when we collected all the staker data
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
	err := w.EntryWritingFunction() // Write to blockchain
	if err != nil {
		log.WithError(err).Error("Failed to write staking record")
		return
	}

	dbht := int32(-1)
	if w.sprTemplate != nil {
		dbht = w.sprTemplate.Dbht
	}

	log.WithFields(log.Fields{
		"height": dbht,
	}).Info("SPR(s) Staked")
}

// writeStakingRecord writes an spr to the blockchain
func (w *EntryWriter) writeStakingRecord() error {
	if w.sprTemplate == nil {
		return fmt.Errorf("no spr template")
	}

	fctAddresses, err := w.config.String("Staker.CoinbaseAddress")
	if err != nil { // Not likely to happen since we
		return errors.New("No fctAddress found") // check for bad addresses earlier
	}
	fctAddrs := strings.Split(fctAddresses, ",")
	countOfRecords := 0
	for _, addr := range fctAddrs {
		operation := func() error {
			var err1, err2 error

			network, _ := common.LoadConfigNetwork(w.config)

			if w.sprTemplate.CoinbasePEGAddress, err =
				common.ConvertFCTtoPegNetAsset(network, "PEG", addr); err != nil {
				log.Errorf("invalid fct address in config file: %v", err)
			}

			w.sprTemplate.CoinbaseAddress = addr

			entry, err := w.sprTemplate.CreateSPREntry()
			switch {
			case err != nil:
				return err
			case entry == nil:
				return errors.New("w.sprTemplate.CreateSPREntry returned a nil entry")
			case len(w.sprTemplate.CoinbaseAddress) == 0:
				return errors.New("w.sprTemplate.CoinbaseAddress is missing")
			case entry.Content == nil:
				return errors.New("entry.Content is nil")
			case entry.ExtIDs == nil:
				return errors.New("entry.ExtIDs is nil")
			case len(entry.ChainID) == 0:
				return errors.New("entry.ChainID is missing")
			default:
				_, err1 = factom.CommitEntry(entry, w.ec)
				_, err2 = factom.RevealEntry(entry)
				if err1 == nil && err2 == nil {
					return nil
				}
			}
			return errors.New("failed to write SPR Entry")
		}
		err = backoff.Retry(operation, common.PegExponentialBackOff())
		if err != nil {
			log.WithFields(log.Fields{
				"error":   err,
				"address": addr,
			}).Error("error encountered while attempting create an SPR")
		} else {
			countOfRecords++
		}
	}
	ecAddress := w.ecUser
	if err != nil {
		panic("entry credit address is invalid: " + err.Error())
	}
	bal, err := factom.GetECBalance(ecAddress)
	if err != nil {
		panic(fmt.Sprintf("entry credit address [%s] is invalid: %s", ecAddress, err.Error()))
	}
	if bal == 0 {
		panic("EC Balance is zero for " + ecAddress)
	}

	timeLeft := float64(bal) / 144 / float64(countOfRecords)
	days := int64(timeLeft)
	hours := int64((timeLeft - float64(days)) * 24)
	minutes := int64(((timeLeft-float64(days))*24 - float64(hours)) * 60)
	log.WithFields(log.Fields{
		"ecAddress":  ecAddress[:8],
		"balance":    bal,
		"day:hr:min": fmt.Sprintf("%4d:%02d:%02d", days, hours, minutes),
	}).Info("EC (balance) (can stake for day:hr:min)")

	return nil
}

// Cancel will cancel a staker's write. If the staker was stopped, we should not expect his write
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

func NewEntryForwarder(config *config.Config, entryChannel chan *factom.Entry) *EntryForwarder {
	n := new(EntryForwarder)
	n.EntryWriter = NewEntryWriter(config)
	n.entryChannel = entryChannel
	n.EntryWritingFunction = n.forwardStakingRecord
	return n
}

// ECBalance is always positive, the coordinator will stop us staking if he runs out
func (w *EntryForwarder) ECBalance() (int64, error) {
	return 1, nil
}

// NextBlockWriter gets the next block writer to use for the staker.
//	Because all stakers will share a block writer, we make this call idempotent
func (w *EntryForwarder) NextBlockWriter() IEntryWriter {
	w.Lock()
	defer w.Unlock()
	if w.Next == nil {
		w.Next = NewEntryForwarder(w.config, w.entryChannel)
	}
	return w.Next
}

func (w *EntryForwarder) forwardStakingRecord() error {
	if w.sprTemplate == nil {
		return fmt.Errorf("no spr template")
	}

	entry, err := w.sprTemplate.CreateSPREntry()
	if err != nil {
		return err
	}

	w.entryChannel <- entry
	return nil
}
