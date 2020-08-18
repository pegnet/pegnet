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
			// Version 1 deprecates on block 210330
			if height < V2GradingActivation { // V1 ends at 210330 on MainNet
				return 1
			}
			if height < FloatingPegPriceActivation {
				return 2
			}
			if height < V4HeightActivation {
				return 3
			}
			if height < V20HeightActivation {
				return 4
			}
			return 5 // Latest code version
		},
		TestNetwork: func(height int64) uint8 {
			if height < 96145 { // V1 ends at 96145 on community testnet
				return 1
			}
			// TODO: Find v3 act on testnet
			return 2
		},
	}

	// StakingHeights indicates the SPR version, which dictates the SPR format.
	StakingHeights = map[string]func(height int64) uint8{
		MainNetwork: func(height int64) uint8 {
			return 5 // Latest code version
		},
		TestNetwork: func(height int64) uint8 {
			return 5
		},
	}

	V2GradingActivation int64 = 210330

	// FloatingPegPriceActivation indicates when to place the PEG price into
	// the opr record from the floating exchange price.
	// Estimated to be  Dec 9, 2019, 17:00 UTC
	FloatingPegPriceActivation int64 = 222270

	// V4HeightActivation indicates the activation of additional currencies and ecdsa keys.
	// Estimated to be  Feb 12, 2020, 18:00 UTC
	V4HeightActivation int64 = 231620

	// V20HeightActivation indicates the activation of PegNet 2.0.
	// Estimated to be  Aug 19th 2020 14:00 UTC
	V20HeightActivation int64 = 258796
)

// NetworkActive returns true if the network height is above the activation height.
// If we are below it, the network is not yet active.
func NetworkActive(network string, height int64) bool {
	if min, ok := ActivationHeights[network]; ok {
		return height >= min
	}
	//Not a network we know of? Default to active.
	return true
}

// OPRVersion returns the OPR version for a given height and network.
// If an OPR has a different version, it is invalid. The version dictates the grading
// algo to use and the OPR format.
func OPRVersion(network string, height int64) uint8 {
	return GradingHeights[network](height)
}

// SPRVersion returns the SPR version for a given height and network.
// If an SPR has a different version, it is invalid. The version dictates the SPR format.
func SPRVersion(network string, height int64) uint8 {
	return StakingHeights[network](height)
}

// SetTestingHeight is used for unit test
func SetTestingVersion(version uint8) {
	GradingHeights[UnitTestNetwork] = func(height int64) uint8 {
		return version
	}
}
