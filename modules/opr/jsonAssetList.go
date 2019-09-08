package opr

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Asset struct {
	Name  string
	Value float64
}
type AssetList map[string]float64

// from https://github.com/iancoleman/orderedmap/blob/master/orderedmap.go#L310
func (al AssetList) MarshalJSON() ([]byte, error) {
	if _, ok := al["version"]; !ok {
		return nil, fmt.Errorf("marshaling json must be called through a safe function")
	}
	version := int(al["version"])
	delete(al, "version")

	if version == 0 {
		return nil, fmt.Errorf("version unset for json marshalling")
	}

	var assets []string
	switch version {
	case 1:
		// We need to read the PNT code instead of peg so the hashes match.
		// When we digest this, we should immediately switch it to PEG
		assets = V1Assets
	case 2:
		assets = V2Assets
	}

	s := "{"

	for _, k := range assets {
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
	if len(assets) > 0 {
		s = s[0 : len(s)-1]
	}
	s = s + "}"
	return []byte(s), nil
}
