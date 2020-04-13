package common

import (
	"crypto/ecdsa"
	"github.com/PlatONnetwork/PlatON-Go/common"
	"math/big"
)

type PriAccount struct {
	Priv     *ecdsa.PrivateKey
	Nonce    uint64
	Address  common.Address
	GasPrice *big.Int
}
