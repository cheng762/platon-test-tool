package testcases

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	ethereum "github.com/PlatONnetwork/PlatON-Go"
	"github.com/PlatONnetwork/PlatON-Go/common"
	"github.com/PlatONnetwork/PlatON-Go/common/hexutil"
	"github.com/PlatONnetwork/PlatON-Go/core/types"
	"github.com/PlatONnetwork/PlatON-Go/ethclient"
	"math/big"
	"strconv"
	"strings"
)

type TxParams struct {
	From     common.Address `json:"from"`
	To       common.Address `json:"to"`
	Gas      string         `json:"gas"`
	GasPrice string         `json:"gasPrice"`
	Value    string         `json:"value"`
	Data     string         `json:"data"`
}

type Response struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  string `json:"result"`
	Id      int    `json:"id"`
	Error   struct {
		Code    int32  `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type UnlockResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  bool   `json:"result"`
	Id      int    `json:"id"`
	Error   struct {
		Code    int32  `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func SendRawTransaction(ctx context.Context, client *ethclient.Client, from, to common.Address, value string, data []byte) (common.Hash, error) {
	if len(AccountPool) == 0 {
		parsePkFile()
	}
	v := new(big.Int)
	var err error
	if strings.HasPrefix(value, "0x") {
		bigValue, err := hexutil.DecodeBig(value)
		if err != nil {
			panic(err)
		}
		v = bigValue
	} else {
		v2, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("transfer value to int error.%s", err))
		}
		v.SetInt64(v2)
	}
	acc, ok := AccountPool[from]
	if !ok {
		return common.ZeroHash, fmt.Errorf("private key not found in private key file,addr:%s", from)
	}
	var msg ethereum.CallMsg
	msg.Data = data
	msg.Value = v
	msg.To = &to
	msg.From = from
	gas, err := client.EstimateGas(ctx, msg)
	if err != nil {
		return common.ZeroHash, fmt.Errorf("EstimateGas fail:%v", err)
	}
	gasPrise, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return common.ZeroHash, fmt.Errorf("SuggestGasPrice fail:%v", err)
	}
	nonce, err := client.NonceAt(ctx, from, nil)
	if err != nil {
		return common.ZeroHash, fmt.Errorf("get nonce fail:%v", err)
	}

	if acc.Nonce < nonce {
		acc.Nonce = nonce
	}

	newTx := getSignedTransaction(from, to, v, gas, gasPrise, data, acc.Priv, acc.Nonce)

	if err := client.SendTransaction(ctx, newTx); err != nil {
		panic(err)
	}
	acc.Nonce = acc.Nonce + 1
	return newTx.Hash(), nil
}

func getSignedTransaction(from, to common.Address, value *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, priv *ecdsa.PrivateKey, nonce uint64) *types.Transaction {
	newTx, err := types.SignTx(types.NewTransaction(nonce, to, value, gasLimit, gasPrice, data), types.NewEIP155Signer(new(big.Int).SetInt64(102)), priv)
	if err != nil {
		panic(fmt.Errorf("sign error,%s", err.Error()))
	}
	return newTx
}
