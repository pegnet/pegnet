// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package main

// This test uses sha256 as a fast simulation of LXRHash to collect data about grading with various
// levels of attack against a varied number of "honest" miners.  Since sha256 is roughly 100 to 500
// times faster than LXRHash, the simulation can run quite fast compared to what can be done with the
// actual LXRHash.
//
// LXRHash is a PoW hash, and has about the same odds of finding difficulty solutions as sha256.

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"os"
	"sort"
)

type OPR struct {
	name   string
	record []byte
	hash   []byte
	nonce  int32
	diff   uint64
	value  float64
	grade  float64
}

func (o *OPR) prt() {
	fmt.Printf("%20s diff %16x value %10.6f  grade %20.18f\n", o.name, o.diff, o.value, o.grade)
}

func (o *OPR) mine() {
	for i := 0; i < 1000; i++ {
		o.nonce++
		r := append(o.record, byte(o.nonce>>24), byte(o.nonce>>16), byte(o.nonce>>8), byte(o.nonce))
		h := sha256.Sum256(r)
		o.hash = h[:]
		o.difficulty()
	}
}

func (o *OPR) difficulty() (d uint64) {
	for i := 0; i < 8; i++ {
		d = d<<8 + uint64(o.hash[i])
	}
	if o.diff < d {
		o.diff = d
	}
	return o.diff
}

func randbytes() (bs []byte) {
	for len(bs) < 100 {
		bs = append(bs, byte(rand.Int()))
	}
	return
}

func addMine(name string, percent float64, oprs []*OPR) {
	for _, opr := range oprs {
		opr.mine()
		opr.name = name
		opr.value = opr.value + opr.value*percent
	}
}

func genOPR(num int) (oprs []*OPR) {
	for len(oprs) < num {
		opr := new(OPR)
		opr.name = "nice"
		opr.record = randbytes()
		opr.value = 5 + float64(rand.Int31n(10000))/10000
		oprs = append(oprs, opr)
	}
	return
}

func ReduceTo50(oprs []*OPR) (best50 []*OPR) {
	var keep []*OPR

	for len(oprs) > 50 {
		for len(oprs)+len(keep) > 50 && len(oprs) > 2 {
			if oprs[0].diff > oprs[1].diff {
				keep = append(keep, oprs[0])
			} else {
				keep = append(keep, oprs[1])
			}
			oprs = oprs[2:]
		}
		oprs = append(oprs, keep...)
		keep = keep[:0]
	}

	return oprs
}

func ReduceTo10(oprs []*OPR) (best10 []*OPR) {
	for len(oprs) > 10 {
		var avg float64
		for _, opr := range oprs {
			avg = avg + opr.value
		}
		avg = avg / float64(len(oprs))

		for _, opr := range oprs {
			d := (avg - opr.value) / avg
			opr.grade = d * d * d * d
		}

		sort.Slice(oprs, func(i, j int) bool { return oprs[i].grade < oprs[j].grade })

		//for _, opr := range oprs {
		//opr.prt()
		//}

		oprs = oprs[:len(oprs)-1]
	}
	return oprs
}

func main() {

	out, err := os.Create("out.txt")
	if err != nil {
		panic(err)
	}

	for weight := 1; weight < 10; weight++ { // How much weight does the attacker have relative to other miners
		for attackRecs := 10; attackRecs < 50; attackRecs++ { // How many records is the attacker going to submit?
			for numMiners := 50; numMiners < 150; numMiners += 10 { // number of miners
				for bias := float64(-.25); bias <= .25; bias += .01 {
					oprs := genOPR(numMiners)

					addMine("bad", bias, oprs[:attackRecs])
					for i := 1; i < weight; i++ {
						addMine("bad", 0, oprs[:weight])
					}

					for i := 0; i < 5; i++ {
						for i, _ := range oprs {
							j := rand.Int31n(int32(len(oprs)))
							oprs[i], oprs[j] = oprs[j], oprs[i]
						}
					}
					for _, opr := range oprs {
						opr.mine()
					}

					badcnt := 0
					for _, opr := range oprs {
						if opr.name == "bad" {
							badcnt++
						}
					}

					oprs = ReduceTo50(oprs)

					badcnt50 := 0
					for _, opr := range oprs {
						if opr.name == "bad" {
							badcnt50++
						}
					}

					oprs = ReduceTo10(oprs)

					badcnt10 := 0
					for _, opr := range oprs {
						if opr.name == "bad" {
							badcnt10++
						}
					}

					_, err := fmt.Fprintf(out, "   wt, %4d,     bias, %5.2f,   #aRec, %4d,    numMiners %4d,    badcnt50, %4d,    badcnt10, %4d,    winner, %s\n",
						weight,
						bias,
						attackRecs,
						numMiners,
						badcnt50,
						badcnt10,
						oprs[0].name)
					if err != nil {
						panic(err)
					}
				}
			}
		}
	}
}
