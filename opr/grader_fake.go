// Copyright (c) of parts are held by the various contributors (see the CLA)
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package opr

// FakeGrader can be used in unit tests
type FakeGrader struct {
	*Grader
}

func NewFakeGrader() *FakeGrader {
	f := new(FakeGrader)
	f.Grader = NewGrader()

	return f
}

func (f *FakeGrader) EmitFakeEvent(event OPRs) {
	for _, a := range f.alerts {
		select { // Don't block if someone isn't pulling from the winner channel
		case a <- &event:
		default:
			// This means the channel is full
		}
	}
}
