package opr

import (
	"encoding/json"
	"strings"
)

type Asset struct {
	Name  string
	Value float64
}
type AssetList map[string]float64

// MarshalJSON marshals a golang map in a consistent order
// implemented from https://github.com/iancoleman/orderedmap/blob/master/orderedmap.go#L310
func (al AssetList) MarshalJSON() ([]byte, error) {
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
