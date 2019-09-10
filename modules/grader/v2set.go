package grader

// V2GradedBlock is an opr set that has been graded. The set should be read only through it's interface
// implementation.
type V2GradedBlock struct {
	*baseGradedBlock
}

func NewV2GradedBlock(graded []*GradingOPR, gradedTo int, height int32) *V2GradedBlock {
	b := new(V2GradedBlock)
	b.baseGradedBlock = newBaseGradedBlock(graded, gradedTo, height, 25)

	return b
}

func (s *V2GradedBlock) Version() uint8 {
	return 1
}
