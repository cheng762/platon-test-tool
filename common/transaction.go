package common

import (
	"fmt"
	"github.com/PlatONnetwork/PlatON-Go/common"
	"github.com/PlatONnetwork/PlatON-Go/core/types"
	"math/big"
)

func getSignedTransaction(from *PriAccount, to common.Address, value *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *types.Transaction {
	newTx, err := types.SignTx(types.NewTransaction(from.Nonce, to, value, gasLimit, gasPrice, data), types.NewEIP155Signer(new(big.Int).SetInt64(102)), from.Priv)
	if err != nil {
		panic(fmt.Errorf("sign error,%s", err.Error()))
	}
	return newTx
}
