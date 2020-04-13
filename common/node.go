package common

import (
	"bytes"
	"context"
	"fmt"
	ethereum "github.com/PlatONnetwork/PlatON-Go"
	"github.com/PlatONnetwork/PlatON-Go/common"
	"github.com/PlatONnetwork/PlatON-Go/common/hexutil"
	"github.com/PlatONnetwork/PlatON-Go/core/types"
	contract "github.com/PlatONnetwork/PlatON-Go/core/vm"
	"github.com/PlatONnetwork/PlatON-Go/ethclient"
	"github.com/PlatONnetwork/PlatON-Go/p2p/discover"
	"github.com/PlatONnetwork/PlatON-Go/rlp"
	"github.com/PlatONnetwork/platon-test-tool/Dto"
	"math/big"
	"strings"
)

type TxManager struct {
	nodes map[string]*node
}

func (txm *TxManager) AddNode(url string) *node {
	node := NewNode(url)
	txm.nodes[url] = node
	return node
}

func (txm *TxManager) GetNode(url string) *node {
	node, ok := txm.nodes[url]
	if !ok {
		return txm.AddNode(url)
	}
	return node
}

func NewNode(url string) *node {
	n := new(node)
	client, err := ethclient.Dial(url)
	if err != nil {
		panic(err)
	}
	n.client = client
	n.ctx = context.Background()
	n.txs = make([]*types.Transaction, 0)
	return n
}

type node struct {
	ctx    context.Context
	client *ethclient.Client
	txs    []*types.Transaction
}

//创建质押
func (n *node) CreateStakingTransaction(from *PriAccount, input Dto.Staking) (*types.Transaction, error) {
	send, to := n.encodePPOS(contract.TxCreateStaking, input.Typ, input.BenefitAddress, input.NodeId, input.ExternalId, input.NodeName, input.Website, input.Details, input.Amount, input.RewardPer, input.ProgramVersion, input.ProgramVersionSign, input.BlsPubKey, input.BlsProof)
	return n.SignedTransaction(from, to, "0", send)
}

//解除质押
func (n *node) WithdrewStaking(from *PriAccount, nodeID discover.NodeID) (*types.Transaction, error) {
	send, to := n.encodePPOS(contract.TxWithdrewCandidate, nodeID)
	return n.SignedTransaction(from, to, "0", send)
}

func (n *node) SendTraction(from *PriAccount, to common.Address, value string, data []byte) (common.Hash, error) {
	tx, err := n.SignedTransaction(from, to, value, data)
	if err != nil {
		return common.ZeroHash, err
	}
	if err := n.client.SendTransaction(n.ctx, tx); err != nil {
		return common.ZeroHash, err
	}
	return tx.Hash(), nil
}

func (n *node) SendRawTraction(txs ...*types.Transaction) error {
	for _, tx := range txs {
		if err := n.client.SendTransaction(n.ctx, tx); err != nil {
			return err
		}
	}
	return nil
}

//will not exec before send
func (n *node) AddTraction(from *PriAccount, to common.Address, value string, data []byte) (common.Hash, error) {
	tx, err := n.SignedTransaction(from, to, value, data)
	if err != nil {
		return common.ZeroHash, err
	}
	n.txs = append(n.txs, tx)
	return tx.Hash(), nil
}

func (n *node) AddRawTraction(txs ...*types.Transaction) {
	for _, tx := range txs {
		n.txs = append(n.txs, tx)
	}
}

func (n *node) SendAllTraction() error {
	for _, tx := range n.txs {
		if err := n.client.SendTransaction(n.ctx, tx); err != nil {
			return err
		}
	}
	return nil
}

func (n *node) SignedTransaction(from *PriAccount, to common.Address, value string, data []byte) (*types.Transaction, error) {
	v := new(big.Int)
	var err error
	if strings.HasPrefix(value, "0x") {
		bigValue, err := hexutil.DecodeBig(value)
		if err != nil {
			panic(err)
		}
		v = bigValue
	} else {
		tmp, ok := new(big.Int).SetString(value, 10)
		if !ok {
			panic(fmt.Sprintf("transfer value to int error.%s", err))
		}
		v = tmp
	}
	var msg ethereum.CallMsg
	msg.Data = data
	msg.Value = v
	msg.To = &to
	msg.From = from.Address
	//todo gas cal local
	gas, err := n.client.EstimateGas(n.ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("EstimateGas fail:%v", err)
	}
	gasPrice := new(big.Int)
	if from.GasPrice == nil {
		gasPrise, err := n.client.SuggestGasPrice(n.ctx)
		if err != nil {
			return nil, fmt.Errorf("SuggestGasPrice fail:%v", err)
		}
		gasPrice.Set(gasPrise)
	} else {
		gasPrice.Set(from.GasPrice)
	}

	nonce, err := n.client.NonceAt(n.ctx, from.Address, nil)
	if err != nil {
		return nil, fmt.Errorf("get nonce fail:%v", err)
	}

	if from.Nonce < nonce {
		from.Nonce = nonce
	}

	newTx := getSignedTransaction(from, to, v, gas+500000, gasPrice.Add(gasPrice, big.NewInt(6000000)), data)
	from.Nonce = from.Nonce + 1
	return newTx, nil
}

func (n *node) encodePPOS(funcType uint16, params ...interface{}) ([]byte, common.Address) {
	par := buildParams(funcType, params...)
	buf := new(bytes.Buffer)
	err := rlp.Encode(buf, par)
	if err != nil {
		panic(fmt.Errorf("encode rlp data fail: %v", err))
	}
	return buf.Bytes(), funcTypeToContractAddress(funcType)
}
