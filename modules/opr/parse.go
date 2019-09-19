package opr

import (
	"encoding/json"
	"fmt"
)

// Parse takes any input and attempts to automatically determine the version
func Parse(data []byte) (OPR, error) {
	p, err := ParseV2Content(data)
	if err == nil {
		return p, nil
	}

	js, err := ParseV1Content(data)
	if err == nil {
		return js, nil
	}

	return nil, fmt.Errorf("unable to detect format")
}

// ParseV1Content parses JSON
func ParseV1Content(data []byte) (*V1Content, error) {
	js := new(V1Content)
	err := json.Unmarshal(data, js)
	if err != nil {
		return nil, err
	}
	return js, nil
}

// ParseV2Content parses Protobuf
func ParseV2Content(data []byte) (*V2Content, error) {
	// Length 0 does not pass an error in the unmarshal, but length 0 contents
	// for an entry is not unreasonable to expect and is obviously incorrect.
	if len(data) == 0 {
		return nil, fmt.Errorf("no bytes to decode")
	}

	proto := new(V2Content)
	err := proto.Unmarshal(data)
	if err != nil {
		return nil, err
	}
	return proto, nil
}
