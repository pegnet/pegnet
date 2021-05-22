package api

import (
	"encoding/json"
	"github.com/pegnet/pegnet/opr"
)

func MapToObject(source interface{}, dst interface{}) error {
	b, err := json.Marshal(source)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}

// -------------------------------------------------------------
// Requests and Parameters

// PostRequest struct to deserialise from request body
type PostRequest struct {
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

// Parameters contains all possible json inputs
// TODO: move into separate parameter structs depending on the request type
type GenericParameters struct {
	Address   *string `json:"address,omitempty"`
	Height    *int64  `json:"height,omitempty"`
	DigitalID string  `json:"miner_id,omitempty"`
	Hash      string  `json:"hash,omitempty"`
}

type PerformanceParameters struct {
	BlockRange BlockRange `json:"block_range"`
	DigitalID  string     `json:"miner_id,omitempty"`
}

// -------------------------------------------------------------
// Responses

// PostResponse to either contain a valid result or error
type PostResponse struct {
	Res interface{} `json:"result"`
	Err *Error      `json:"error"`
}

// Result struct contains all potential json api responses
// TODO: move into separate parameter structs depending on the response type
type GenericResult struct {
	Balance      int64                         `json:"balance,omitempty"`
	Balances     map[string]map[[32]byte]int64 `json:"balances,omitempty"`
	ChainID      string                        `json:"chain_id,omitempty"`
	LeaderHeight int64                         `json:"leader_height,omitempty"`
	OPRBlocks    []*opr.OprBlock               `json:"opr_blocks,omitempty"`
	Winners      []string                      `json:"winners,omitempty"`
	Winner       string                        `json:"winner,omitempty"`
	OPRBlock     *opr.OprBlock                 `json:"opr_block,omitempty"`
	OPRs         []opr.OraclePriceRecord       `json:"oprs,omitempty"`
	OPR          *opr.OraclePriceRecord        `json:"opr,omitempty"`
}

type PerformanceResult struct {
	BlockRange           BlockRange       `json:"block_range"`
	Submissions          int64            `json:"submissions"`
	Rewards              int64            `json:"rewards"`
	DifficultyPlacements map[string]int64 `json:"difficulty_placements"`
	GradingPlacements    map[string]int64 `json:"grading_placements"`
}

// -------------------------------------------------------------
// Miscellaneous helper structs that appear in both requests and responses

// BlockRange specifies a range of directory blocks with a start and end height (inclusive)
// Negative numbers indicate a request for that many blocks behind the current head
type BlockRange struct {
	Start *int64 `json:"start"`
	End   *int64 `json:"end"`
}
