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
	if err == nil {
		return js, nil
	}
	return js, nil
}

// ParseV2Content parses Protobuf
func ParseV2Content(data []byte) (*V2Content, error) {
	proto := new(V2Content)
	err := proto.Unmarshal(data)
	if err != nil {
		return nil, err
	}
	return proto, nil
}
