package grader

// ValidateV3 is a wrapper for ValidateV2 and additionally checks if PEG > 0
func ValidateV3(entryhash []byte, extids [][]byte, height int32, winners []string, content []byte) (*GradingOPR, error) {
	opr, err := ValidateV2(entryhash, extids, height, winners, content)
	if err != nil {
		return nil, err
	}

	if opr.OPR.GetOrderedAssetsUint()[0].Value == 0 { // length of assets checked in v3
		return nil, NewValidateError("assets must be greater than 0")
	}

	return opr, nil
}
