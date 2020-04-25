package common

import (
	"crypto/ecdsa"
	"fmt"
	ethereum "github.com/PlatONnetwork/PlatON-Go"
	"github.com/PlatONnetwork/PlatON-Go/common"
	"github.com/PlatONnetwork/PlatON-Go/common/hexutil"
	"github.com/PlatONnetwork/PlatON-Go/core/types"
	contract "github.com/PlatONnetwork/PlatON-Go/core/vm"
	platonNode "github.com/PlatONnetwork/PlatON-Go/node"
	"github.com/PlatONnetwork/PlatON-Go/p2p/discover"
	"github.com/PlatONnetwork/platon-test-tool/Dto"
	"math/big"
	"strings"
)

type PriAccount struct {
	Priv     *ecdsa.PrivateKey
	Nonce    uint64
	Address  common.Address
	GasPrice *big.Int
}

//创建质押
func (p *PriAccount) CreateStakingTransaction(node *node, input Dto.Staking) (*types.Transaction, error) {
	send, to := encodePPOS(contract.TxCreateStaking, input.Typ, input.BenefitAddress, input.NodeId, input.ExternalId, input.NodeName, input.Website, input.Details, input.Amount, input.RewardPer, input.ProgramVersion, input.ProgramVersionSign, input.BlsPubKey, input.BlsProof)
	return p.SignedTransaction(node, to, "0", send)
}

//解除质押
func (p *PriAccount) WithdrewStaking(node *node, nodeID discover.NodeID) (*types.Transaction, error) {
	send, to := encodePPOS(contract.TxWithdrewCandidate, nodeID)
	return p.SignedTransaction(node, to, "0", send)
}

func (p *PriAccount) WithdrawDelegateReward(node *node) (*types.Transaction, error) {
	send, to := encodePPOS(contract.TxWithdrawDelegateReward)
	return p.SignedTransaction(node, to, "0", send)
}

func (p *PriAccount) DeclareVersion(node *node, nodePrikey *ecdsa.PrivateKey, activeNode discover.NodeID, programVersion uint32) (*types.Transaction, error) {
	handle := platonNode.GetCryptoHandler()
	handle.SetPrivateKey(nodePrikey)
	versionSign := common.VersionSign{}
	versionSign.SetBytes(handle.MustSign(programVersion))
	send, to := encodePPOS(contract.Declare, activeNode, programVersion, versionSign)
	return p.SignedTransaction(node, to, "0", send)
}

//创建委托
func (p *PriAccount) DelegateTransaction(node *node, nodeID discover.NodeID, typ uint16, amount *big.Int) (*types.Transaction, error) {
	send, to := encodePPOS(contract.TxDelegate, typ, nodeID, amount)
	return p.SignedTransaction(node, to, "0", send)
}

//减持/撤销委托
func (p *PriAccount) WithdrewDelegateTransaction(node *node, stakingBlockNum uint64, nodeID discover.NodeID, amount *big.Int) (*types.Transaction, error) {
	send, to := encodePPOS(contract.TxWithdrewDelegate, stakingBlockNum, nodeID, amount)
	return p.SignedTransaction(node, to, "0", send)
}

func (p *PriAccount) SignedTransaction(node *node, to common.Address, value string, data []byte) (*types.Transaction, error) {
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
	msg.From = p.Address
	//todo gas cal local
	gas, err := node.client.EstimateGas(node.ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("EstimateGas fail:%v", err)
	}
	gasPrice := new(big.Int)
	if p.GasPrice == nil {
		gasPrise, err := node.client.SuggestGasPrice(node.ctx)
		if err != nil {
			return nil, fmt.Errorf("SuggestGasPrice fail:%v", err)
		}
		gasPrice.Set(gasPrise)
	} else {
		gasPrice.Set(p.GasPrice)
	}

	nonce, err := node.client.NonceAt(node.ctx, p.Address, nil)
	if err != nil {
		return nil, fmt.Errorf("get nonce fail:%v", err)
	}

	if p.Nonce < nonce {
		p.Nonce = nonce
	}

	newTx := getSignedTransaction(p, to, v, gas+500000, gasPrice.Add(gasPrice, big.NewInt(6000000)), data)
	p.Nonce = p.Nonce + 1
	return newTx, nil
}
