package src

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/pegnet/pegnet/opr"
	"github.com/spf13/cobra"
	"go.uber.org/ratelimit"
)

func init() {
	rootCmd.AddCommand(best)
	rootCmd.AddCommand(blocks)

	blocks.Flags().Int("blocks", 100, "How many 'blocks' to create.")
	blocks.Flags().IntP("target", "t", 300, "How many OPRs to target")
	blocks.Flags().String("csv", "", "Decide to write stats to a csv")
}

var HashRateLimit ratelimit.Limiter

type PerBlock struct {
	Height          int
	Duration        time.Duration
	TotalHashes     uint64
	NonceAggregator *opr.NonceRanking
}

func (b *PerBlock) CsvHeader() []string {
	return []string{"BlockHeight", "Records",
		"Record Target", "MinDifficulty (int)", "MinDifficulty (hex)",
		"Best Diff (int)", "Best Diff (hex)",
		"Last Graded Index", "Last Graded (int)", "Last Graded (hex)",
		"Last Index", "Last (int)", "Last (hex)",
		"Total Hashes", "Hash Rate"}
}

func (b *PerBlock) Records(target int, minDiff uint64) []string {
	list := b.NonceAggregator.GetNonces()
	lastGradedSpot, lastGradedDiff := DifficultyAtFromList(50, list)

	return []string{
		fmt.Sprintf("%d", b.Height),
		fmt.Sprintf("%d", len(list)),
		fmt.Sprintf("%d", target),
		fmt.Sprintf("%d", minDiff),
		fmt.Sprintf("%x", minDiff),

		// best
		fmt.Sprintf("%d", list[0].Difficulty),
		fmt.Sprintf("%x", list[0].Difficulty),

		// Last graded
		fmt.Sprintf("%d", lastGradedSpot),
		fmt.Sprintf("%d", lastGradedDiff),
		fmt.Sprintf("%x", lastGradedDiff),

		// Last
		fmt.Sprintf("%d", len(list)),
		fmt.Sprintf("%d", list[len(list)-1].Difficulty),
		fmt.Sprintf("%x", list[len(list)-1].Difficulty),

		fmt.Sprintf("%d", b.TotalHashes),
		fmt.Sprintf("%.3f", float64(b.TotalHashes)/b.Duration.Seconds()),
	}
}

func (b *PerBlock) String() string {
	return fmt.Sprintf("Height %2d, Records %3d, TotalHashes: %5d, Total HashRate: %.3f,",
		b.Height,
		len(b.NonceAggregator.GetNonces()),
		b.TotalHashes,
		float64(b.TotalHashes)/b.Duration.Seconds(),
	)
}

// DifficultyAt will return the difficulty at the spot or the closest spot to it
// if we don't have enough. So if you target 50, and we only have 4, it will return the 4th.
func (b *PerBlock) DifficultyAt(spot int) (index int, diff uint64) {
	list := b.NonceAggregator.GetNonces()
	return DifficultyAtFromList(spot, list)
}

// DifficultyAt will return the difficulty at the spot or the closest spot to it
// if we don't have enough. So if you target 50, and we only have 4, it will return the 4th.
func DifficultyAtFromList(spot int, list []*opr.UniqueOPRData) (index int, diff uint64) {
	if len(list) >= spot {
		// Spot is 1 based
		return spot, list[spot-1].Difficulty
	}

	return len(list), list[len(list)-1].Difficulty
}

var blocks = &cobra.Command{
	Use:   "blocks",
	Short: "Simulate mining blocks, and reports the 1st and 50th difficulty. As well as how many records were created.",
	Long: "This tests the calculation to dial the number of OPRs per block. The `target` is what we are aiming for. the first " +
		"block will always have 50 oprs, as we need to bootstrap the process.",
	Run: func(cmd *cobra.Command, args []string) {
		csvPath, _ := cmd.Flags().GetString("csv")
		var writer *csv.Writer
		if csvPath != "" {
			var _ = os.Remove(csvPath)
			file, err := os.OpenFile(csvPath, os.O_CREATE|os.O_RDWR, 0666)
			if err != nil {
				panic(err)
			}
			writer = csv.NewWriter(file)
			tmp := new(PerBlock)
			var _ = writer.Write(tmp.CsvHeader())
			writer.Flush()

			defer writer.Flush()
			defer file.Close()
		}

		blocks, _ := cmd.Flags().GetInt("blocks")
		target, _ := cmd.Flags().GetInt("target")
		v, _ := cmd.Flags().GetString("time")
		d, err := time.ParseDuration(v)
		if err != nil {
			panic(err)
		}

		allBlocks := make([]*PerBlock, blocks)

		for i := 0; i < blocks; i++ {
			// If we added flux, this will adjust the hashrate
			UpdateHashRate(cmd, args)
			minDiff := uint64(0)
			b := new(PerBlock)
			b.Height = i
			if i == 0 {
				// 50 for the first block
				b.NonceAggregator = opr.NewNonceRanking(50)
			} else {
				// 1000 is upper bound
				b.NonceAggregator = opr.NewNonceRanking(1000)
				// Target 'target' based on the 50th of the prior
				prior := allBlocks[i-1]
				spot, diff := prior.DifficultyAt(50)
				minDiff = opr.CalculateMinimumDifficulty(spot, diff, target)
				b.NonceAggregator.MinimumDifficulty = minDiff
			}

			// Run for duration d and find the best hash
			timer := time.NewTimer(d)
		BlockLoop:
			for {
				select {
				case <-timer.C:
					break BlockLoop
				default:
					tmp := RandomDifficulty()
					b.NonceAggregator.AddNonce([]byte{}, tmp)
					b.TotalHashes++
				}
			}
			b.Duration = d
			allBlocks[i] = b
			fmt.Println(b)

			if writer != nil {
				var _ = writer.Write(b.Records(target, minDiff))
				writer.Flush()
			}
		}
	},
}

var best = &cobra.Command{
	Use:   "best",
	Short: "simulate running for X seconds and report the highest found",
	Run: func(cmd *cobra.Command, args []string) {
		v, _ := cmd.Flags().GetString("time")
		d, err := time.ParseDuration(v)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Running for %s...\n", d)

		hashes := uint64(0)
		best := uint64(0)

		// Run for duration d and find the best hash
		timer := time.NewTimer(d)

	OuterLoop:
		for {
			select {
			case <-timer.C:
				break OuterLoop
			default:
				tmp := RandomDifficulty()
				if tmp > best {
					best = tmp
				}
				hashes++
			}
		}

		fmt.Printf("Best difficulty found: %x | %d\n", best, best)
		fmt.Printf("Total hashes: %d\n", hashes)
		fmt.Printf("Hashrate: %.3f/s\n", float64(hashes)/d.Seconds())
	},
}

func RandomDifficulty() uint64 {
	if HashRateLimit != nil {
		HashRateLimit.Take()
	}
	buf := make([]byte, 8)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}

	return binary.BigEndian.Uint64(buf)
}
