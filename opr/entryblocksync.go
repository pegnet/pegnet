package opr

import (
	"github.com/FactomProject/factom"
	"github.com/pegnet/pegnet/common"
)

// EntryBlockSync has the current eblock synced to, and the target Eblock
//	It also has the blocks in between in order. This makes it so traversing only needs to happen once
type EntryBlockSync struct {
	ChainID          string             // Chainid of the eblock chain
	Current          EntryBlockMarker   // The current eblock we have synced
	Target           EntryBlockMarker   // The target is the chainhead
	BlocksToBeParsed []EntryBlockMarker // The eblocks between the current and target (including target)
}

func NewEntryBlockSync(chainid string) *EntryBlockSync {
	e := new(EntryBlockSync)
	e.ChainID = chainid

	return e
}

// SyncBlocks will query factomd and check if the chainhead has been updated,
// and fill in the non synced eblocks
func (a *EntryBlockSync) SyncBlocks() error {
	// First check to see if the chainhead has been updated
	expChainHead := a.Head()

	heb, _, err := factom.GetChainHead(a.ChainID)
	if err != nil {
		return common.DetailError(err)
	}

	if expChainHead.KeyMr != heb {
		// We have a new chainhead. This means we need to walk backwards in our entryblocks
		// until we find our target. And add all the eblocks we are missing.
		// From there we will have a clean Eblock sync to properly sync from, and use eblocks
		// as checkpoints.
		var eblocks []EntryBlockMarker
		next := heb
		for {
			// If we found the target, we can stop
			// Or if we found the first eblock in the chain, we can stop.
			if next == expChainHead.KeyMr || next == common.ZeroHash {
				break
			}

			eblock, err := factom.GetEBlock(next)
			if err != nil {
				return err
			}

			eblocks = append(eblocks, EntryBlockMarker{next, eblock})
			next = eblock.Header.PrevKeyMR
		}

		// Add the blocks to our sync in the order they are on chain
		for i := len(eblocks) - 1; i >= 0; i-- {
			a.AddNewHeadMarker(eblocks[i])
		}
	}
	return nil
}

// Synced returns if fully synced (current == target)
func (a *EntryBlockSync) Synced() bool {
	return a.Current.IsSameAs(&a.Target)
}

// Head returns our target head. This is the chainhead that we know of
func (a *EntryBlockSync) Head() EntryBlockMarker {
	return a.Target
}

// NextEBlock returns the next eblock that is needed to be parsed
func (a *EntryBlockSync) NextEBlock() *EntryBlockMarker {
	if len(a.BlocksToBeParsed) == 0 {
		return nil
	}
	return &a.BlocksToBeParsed[0]
}

// BlockParsed indicates a block has been parsed. We update our current
func (a *EntryBlockSync) BlockParsed(block EntryBlockMarker) {
	if !a.BlocksToBeParsed[0].IsSameAs(&block) {
		panic("This block should not be next in the list")
	}
	a.Current = block
	tmp := make([]EntryBlockMarker, len(a.BlocksToBeParsed)-1)
	copy(tmp, a.BlocksToBeParsed[1:])
	a.BlocksToBeParsed = tmp
}

func (a *EntryBlockSync) AddNewHead(keymr string, eblock *factom.EBlock) {
	a.AddNewHeadMarker(EntryBlockMarker{keymr, eblock})
}

// AddNewHead will add a new eblock to be parsed to the head (tail of list)
//	Since the block needs to be parsed, it is the new target and added to the blocks to be parsed
func (a *EntryBlockSync) AddNewHeadMarker(marker EntryBlockMarker) {
	if a.Target.KeyMr != "" && marker.EntryBlock.Header.BlockSequenceNumber < a.Target.EntryBlock.Header.BlockSequenceNumber {
		return // Already added this target
	}
	a.BlocksToBeParsed = append(a.BlocksToBeParsed, marker)
	a.Target = marker
}

func (a *EntryBlockSync) IsSameAs(b *EntryBlockSync) bool {
	if !a.Current.IsSameAs(&b.Current) {
		return false
	}
	if !a.Target.IsSameAs(&b.Target) {
		return false
	}

	if len(a.BlocksToBeParsed) != len(b.BlocksToBeParsed) {
		return false
	}

	for i := range a.BlocksToBeParsed {
		if !a.BlocksToBeParsed[i].IsSameAs(&b.BlocksToBeParsed[i]) {
			return false
		}
	}

	return true
}

type EntryBlockMarkerList []EntryBlockMarker

func (p EntryBlockMarkerList) Len() int {
	return len(p)
}

func (p EntryBlockMarkerList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p EntryBlockMarkerList) Less(i, j int) bool {
	return p[i].EntryBlock.Header.BlockSequenceNumber < p[j].EntryBlock.Header.BlockSequenceNumber
}

type EntryBlockMarker struct {
	KeyMr      string
	EntryBlock *factom.EBlock
}

func NewEntryBlockMarker() *EntryBlockMarker {
	e := new(EntryBlockMarker)
	return e
}

func (a *EntryBlockMarker) IsSameAs(b *EntryBlockMarker) bool {
	// Keymrs should be uniq to the eblock
	if a.KeyMr != b.KeyMr {
		return false
	}
	return true
}
