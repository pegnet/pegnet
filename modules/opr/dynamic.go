package opr

import (
	"encoding/json"
	"fmt"
)

func DynamicUnmarshal(data []byte) (OPR, error) {
	p, err := ParseProtobuf(data)
	if err == nil {
		return p, nil
	}

	js, err := ParseJSON(data)
	if err == nil {
		return js, nil
	}

	return nil, fmt.Errorf("unable to detect format")
}

func ParseJSON(data []byte) (*JSONOPR, error) {
	js := new(JSONOPR)
	err := json.Unmarshal(data, js)
	if err == nil {
		return js, nil
	}
	return js, nil
}

func ParseProtobuf(data []byte) (*ProtoOPR, error) {
	proto := new(ProtoOPR)
	err := proto.Unmarshal(data)
	if err != nil {
		return nil, err
	}
	return proto, nil
}
