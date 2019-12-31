package main

import (
	"fmt"

	"github.com/FactomProject/factom"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/modules/grader"
	log "github.com/sirupsen/logrus"
)

func main() {
	factom.SetFactomdServer("https://api.factomd.net")
	height := 225338
	err := grade(int64(height))
	if err != nil {
		panic(err)
	}
}

func grade(height int64) error {
	dblock, _, err := factom.GetDBlockByHeight(height)
	if err != nil {
		return err
	}

	var keymr string
	for _, eblock := range dblock.DBEntries {
		if eblock.ChainID == "a642a8674f46696cc47fdb6b65f9c87b2a19c5ea8123b3d2f0c13b6f33a9d5ef" {
			keymr = eblock.KeyMR
			break
		}
	}

	if keymr == "" {
		return fmt.Errorf("no pegnet block found")
	}

	entries, err := getEntries(keymr)
	if err != nil {
		return err
	}

	err = moduleGrade(height, entries)
	if err != nil {
		return err
	}
	return nil
}

type Entry struct {
	EntryHash []byte
	ExtIDs    [][]byte
	Content   []byte
}

func getEntries(keymr string) ([]Entry, error) {
	eblock, err := factom.GetEBlock(keymr)
	if err != nil {
		return nil, err
	}
	var ents []Entry
	for _, ent := range eblock.EntryList {
		e, err := factom.GetEntry(ent.EntryHash)
		if err != nil {
			return nil, err
		}
		ents = append(ents, Entry{
			EntryHash: e.Hash(),
			ExtIDs:    e.ExtIDs,
			Content:   e.Content,
		})
	}
	return ents, nil
}

func moduleGrade(height int64, ents []Entry) error {
	grader.InitLX()
	v := common.OPRVersion(common.MainNetwork, height)
	g, err := grader.NewGrader(v, int32(height), []string{
		"30e4b0ffc46fe5ba",
		"073aaba79d989e16",
		"beb13d76e5b7b58d",
		"b5711112aade5cf4",
		"2e7b909fdab3cba1",
		"187dfc37df893db6",
		"5389ae24d4629ccd",
		"520f9142214a6740",
		"c07f3989be606372",
		"6f4442ee6d7c53f8",
		"12e443a42e2d3626",
		"b36f7307525d21be",
		"14960369d0dde44b",
		"b3c893e02edb75c8",
		"08b56281d73efce2",
		"11958d032d7f25b2",
		"17484280fe5f78db",
		"78721bf33122bd21",
		"a0797593f1717d1a",
		"7881dbc4a97a2c84",
		"269dfee2840ae158",
		"b3f694cc74980ea0",
		"1c1066580ee88847",
		"e317b8187f962a3c",
		"25614907144764a3",
	})
	if err != nil {
		return err
	}

	for _, ent := range ents {
		err := g.AddOPR(ent.EntryHash, ent.ExtIDs, ent.Content)
		if err != nil {
			log.WithFields(log.Fields{
				"hash":  fmt.Sprintf("%x", ent.EntryHash),
				"error": err,
			}).Errorf("failed to add opr")
		}
	}

	graded := g.Grade()
	for i, win := range graded.Winners() {
		fmt.Printf("%d: %x\n", i, win.EntryHash)
	}
	return nil
}
