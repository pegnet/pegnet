package grader

// V4GradedBlock is an opr set that has been graded. The set should be read only through it's interface
// implementation.
type V4GradedBlock struct {
	GradedBlock
}

// Version returns the underlying grader's version
func (g *V4GradedBlock) Version() uint8 {
	return 4
}
