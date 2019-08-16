package common

var (
	ActivationHeights = map[string]int64{
		// roughly 17:00 UTC on Monday 8/19/2019
		MainNetwork: 206422,
		TestNetwork: 0,
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
