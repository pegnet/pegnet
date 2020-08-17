package spr

import (
	"fmt"
	"github.com/pegnet/pegnet/modules/opr"
)

// ParseS1Content parses Protobuf
func ParseS1Content(data []byte) (*opr.V2Content, error) {
	// Length 0 does not pass an error in the unmarshal, but length 0 contents
	// for an entry is not unreasonable to expect and is obviously incorrect.
	if len(data) == 0 {
		return nil, fmt.Errorf("no bytes to decode")
	}

	proto := new(opr.V2Content)
	err := proto.Unmarshal(data)
	if err != nil {
		return nil, err
	}
	return proto, nil
}
