package grader

// V1GradedBlock is an opr set that has been graded. The set should be read only through it's interface
// implementation.
type V1GradedBlock struct {
	*baseGradedBlock
}

func NewV1GradedBlock(graded []*GradingOPR, gradedTo int, height int32) *V1GradedBlock {
	b := new(V1GradedBlock)
	b.baseGradedBlock = newBaseGradedBlock(graded, gradedTo, height, 10)

	return b
}

func (s *V1GradedBlock) Version() uint8 {
	return 1
}
