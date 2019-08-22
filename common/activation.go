package common

var (
	// ActivationHeights signals the network is active and all miners can begin.
	// This is the signal that pegnet has launched.
	ActivationHeights = map[string]int64{
		// roughly 17:00 UTC on Monday 8/19/2019
		MainNetwork: 206422,
		TestNetwork: 0,
	}

	// GradingHeights indicates the OPR version, which dictates the grading and OPR format.
	// When we switch formats. The activation height indicates the block that grading changes.
	// So if the grading change is on block 100, then the entries in block 100 will be using the
	// new grading format.
	GradingHeights = map[string]func(height int64) uint8{
		MainNetwork: func(height int64) uint8 {
			switch {
			// Version 1 deprecates on block XXXXXX
			// TODO: Set a real block height activate height
			case height < 500000:
				return 1
			}
			return 2 // Latest code version
		},
		TestNetwork: func(height int64) uint8 {
			switch {
			case height < 206869:
				return 1
			}
			return 2
		},
	}
)

// NetworkActive returns true if the network height is above the activation height.
// If we are below it, the network is not yet active.
func NetworkActive(network string, height int64) bool {
	if min, ok := ActivationHeights[network]; ok {
		return height >= min
	}
	// Not a network we know of? Default to active.
	return true
}

// OPRVersion returns the OPR version for a given height and network.
// If an OPR has a different version, it is invalid. The version dictates the grading
// algo to use and the OPR format.
func OPRVersion(network string, height int64) uint8 {
	return GradingHeights[network](height)
}
