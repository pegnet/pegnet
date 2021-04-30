package staking

import (
	"fmt"
	"github.com/Factom-Asset-Tokens/factom"
)

var (
	SIRChain = factom.NewBytes32("f44c503def6b0c0e1fcac40f360cba434195321def4e9ca33d3b583d3a0dff62")
)

func GetStakingIdentity(height int32) (string, error) {
	cl := factom.NewClient()
	cl.FactomdServer = "https://api.factomd.net/v2" // Todo: replace server url from config

	heights := new(factom.Heights)
	err := heights.Get(nil, cl)
	if err != nil {
		fmt.Println("factom height is not getting correctly")
	}

	dblock := new(factom.DBlock)
	dblock.Height = uint32(height)

	if err := dblock.Get(nil, cl); err != nil {
		return "", err
	}

	sirEBlock := dblock.EBlock(SIRChain)
	if sirEBlock != nil {
		if err := multiFetch(sirEBlock, cl); err != nil {
			return "", err
		}
	}

	return "", nil
}

func multiFetch(eblock *factom.EBlock, c *factom.Client) error {
	err := eblock.Get(nil, c)
	if err != nil {
		return err
	}

	work := make(chan int, len(eblock.Entries))
	defer close(work)
	errs := make(chan error)
	defer close(errs)

	for i := 0; i < 8; i++ {
		go func() {
			// TODO: Fix the channels such that a write on a closed channel never happens.
			//		For now, just kill the worker go routine
			defer func() {
				recover()
			}()

			for j := range work {
				errs <- eblock.Entries[j].Get(nil, c)
			}
		}()
	}

	for i := range eblock.Entries {
		work <- i
	}

	count := 0
	for e := range errs {
		count++
		if e != nil {
			// If we return, we close the errs channel, and the working go routine will
			// still try to write to it.
			return e
		}
		if count == len(eblock.Entries) {
			break
		}
	}

	return nil
}
