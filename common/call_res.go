package common

import "math/big"

// for plugin test
type RestrictingResult struct {
	Balance *big.Int                 `json:"balance"`
	Debt    *big.Int                 `json:"debt"`
	Entry   []RestrictingReleaseInfo `json:"plans"`
	Pledge  *big.Int                 `json:"Pledge"`
}

// for plugin test
type RestrictingReleaseInfo struct {
	Height uint64   `json:"blockNumber"` // blockNumber representation of the block number at the released epoch
	Amount *big.Int `json:"amount"`      // amount representation of the released amount
}
