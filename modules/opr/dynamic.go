package opr

import (
	"encoding/json"
	"fmt"
)

func DynamicUnmarshal(data []byte) (OPR, error) {
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

func ParseV1Content(data []byte) (*V1Content, error) {
	js := new(V1Content)
	err := json.Unmarshal(data, js)
	if err == nil {
		return js, nil
	}
	return js, nil
}

func ParseV2Content(data []byte) (*V2Content, error) {
	proto := new(V2Content)
	err := proto.Unmarshal(data)
	if err != nil {
		return nil, err
	}
	return proto, nil
}
