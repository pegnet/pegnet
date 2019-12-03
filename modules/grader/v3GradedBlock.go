package grader

// V3GradedBlock is an opr set that has been graded. The set should be read only through it's interface
// implementation.
type V3GradedBlock struct {
	GradedBlock
}

// Version returns the underlying grader's version
func (g *V3GradedBlock) Version() uint8 {
	return 3
}
