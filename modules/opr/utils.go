package opr

// ValidFCTAddress returns if the address is a public FA address
// TODO: Implement this for real
func ValidFCTAddress(addr string) bool {
	return len(addr) > 2 && addr[:2] == "FA"
}
