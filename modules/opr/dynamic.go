package opr

import (
	"encoding/json"
	"fmt"
)

func DynamicUnmarshal(data []byte) (OPR, error) {
	proto := new(ProtoOPR)
	err := proto.Unmarshal(data)
	if err == nil {
		return proto, nil
	}

	js := new(JSONOPR)
	err = json.Unmarshal(data, js)
	if err == nil {
		return js, nil
	}

	return nil, fmt.Errorf("unable to detect format")
}
