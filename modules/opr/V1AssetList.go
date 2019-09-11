package opr

import (
	"encoding/json"
	"strings"
)

// V1AssetList holds the V1 OPR's asset => value association
type V1AssetList map[string]float64

// MarshalJSON marshals a golang map in a consistent order
// implemented from https://github.com/iancoleman/orderedmap/blob/master/orderedmap.go#L310
func (al V1AssetList) MarshalJSON() ([]byte, error) {
	s := "{"

	for _, k := range V1Assets {
		// add key
		esc := strings.Replace(k, `"`, `\"`, -1)
		s = s + `"` + esc + `":`
		// add value
		v := al[k]
		vBytes, err := json.Marshal(v)
		if err != nil {
			return []byte{}, err
		}
		s = s + string(vBytes) + ","
	}
	if len(V1Assets) > 0 {
		s = s[0 : len(s)-1]
	}
	s = s + "}"
	return []byte(s), nil
}

// Clone the AssetList
func (al V1AssetList) Clone() V1AssetList {
	clone := make(V1AssetList)
	for k, v := range al {
		clone[k] = v
	}
	return clone
}
