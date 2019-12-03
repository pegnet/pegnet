package grader

var _ BlockGrader = (*V3BlockGrader)(nil)

// V3BlockGrader implements the V2 grading algorithm but requires PEG to have a price.
type V3BlockGrader struct {
	V2BlockGrader
}

// Version 3
func (v3 *V3BlockGrader) Version() uint8 {
	return 3
}

// AddOPR verifies and adds a V3 OPR.
func (v3 *V3BlockGrader) AddOPR(entryhash []byte, extids [][]byte, content []byte) error {
	gopr, err := ValidateV3(entryhash, extids, v3.height, v3.prevWinners, content)
	if err != nil {
		return err
	}

	v3.oprs = append(v3.oprs, gopr)
	return nil
}
