package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	ethereum "github.com/PlatONnetwork/PlatON-Go"
	"github.com/PlatONnetwork/PlatON-Go/common"
	"github.com/PlatONnetwork/PlatON-Go/core/types"
	contract "github.com/PlatONnetwork/PlatON-Go/core/vm"
	"github.com/PlatONnetwork/PlatON-Go/ethclient"
	"github.com/PlatONnetwork/PlatON-Go/p2p/discover"
	"github.com/PlatONnetwork/PlatON-Go/params"
	"github.com/PlatONnetwork/PlatON-Go/rlp"
	"github.com/PlatONnetwork/PlatON-Go/x/restricting"
	"github.com/PlatONnetwork/PlatON-Go/x/staking"
	"github.com/PlatONnetwork/PlatON-Go/x/xcom"

	"math/big"
)

func NewNode(url string, chainID uint64) *Node {
	n := new(Node)
	client, err := ethclient.Dial(url)
	if err != nil {
		panic(err)
	}
	n.client = client
	n.ctx = context.Background()
	n.txs = make([]*types.Transaction, 0)
	n.id = url
	n.chainID = chainID
	return n
}

type Node struct {
	ctx     context.Context
	client  *ethclient.Client
	txs     []*types.Transaction
	id      string
	chainID uint64
}

func (n *Node) GetID() string {
	return n.id
}

func (n *Node) BlockByNumber(number *big.Int) (*types.Block, error) {
	return n.client.BlockByNumber(n.ctx, number)
}

//获取账户金额
func (n *Node) GetBalance(address common.Address, blockNumber *big.Int) *big.Int {
	balance, err := n.client.BalanceAt(n.ctx, address, blockNumber)
	if err != nil {
		panic(err)
	}
	return balance
}

func (n *Node) GetNonceAt(account common.Address, blockNumber *big.Int) (uint64, error) {
	return n.client.NonceAt(n.ctx, account, blockNumber)
}

//获取ProgramVersion
func (n *Node) GetSchnorrNIZKProve() (string, error) {
	prove, err := n.client.GetSchnorrNIZKProve(n.ctx)
	if err != nil {
		return "", err
	}
	return prove, nil
}

//获取ProgramVersion
func (n *Node) CallProgramVersion() (*params.ProgramVersion, error) {
	pg, err := n.client.GetProgramVersion(n.ctx)
	if err != nil {
		return nil, err
	}
	return pg, nil
}

//will not exec before send
func (n *Node) AddTraction(from *PriAccount, to common.Address, value string, data []byte) (common.Hash, error) {
	tx, err := from.SignedTransaction(n, to, value, data)
	if err != nil {
		return common.ZeroHash, err
	}
	n.txs = append(n.txs, tx)
	return tx.Hash(), nil
}

func (n *Node) AddRawTraction(txs ...*types.Transaction) {
	for _, tx := range txs {
		n.txs = append(n.txs, tx)
	}
}

func (n *Node) SendTraction(from *PriAccount, to common.Address, value string, data []byte) (common.Hash, error) {
	tx, err := from.SignedTransaction(n, to, value, data)
	if err != nil {
		return common.ZeroHash, err
	}
	if err := n.client.SendTransaction(n.ctx, tx); err != nil {
		return common.ZeroHash, err
	}
	return tx.Hash(), nil
}

func (n *Node) SendRawTraction(txs ...*types.Transaction) error {
	for _, tx := range txs {
		if err := n.client.SendTransaction(n.ctx, tx); err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) SendAllTraction() error {
	for _, tx := range n.txs {
		if err := n.client.SendTransaction(n.ctx, tx); err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) CallPPosContract(funcType uint16, params ...interface{}) xcom.Result {
	send, to := encodePPOS(funcType, params...)
	var msg ethereum.CallMsg
	msg.Data = send
	msg.To = &to
	res, err := n.client.CallContract(n.ctx, msg, nil)
	if err != nil {
		panic(err)
	}
	var xres xcom.Result
	if err := json.Unmarshal(res, &xres); err != nil {
		panic(err)
	}
	return xres
}

//获取锁仓信息
func (n *Node) CallGetRestrictingInfo(ctx context.Context, account common.Address) RestrictingResult {
	xres := n.CallPPosContract(4100, account)
	if xres.Code == 0 {
		var result restricting.Result
		result = xres.Ret.(restricting.Result)
		var res RestrictingResult
		res.Balance = result.Balance.ToInt()
		res.Pledge = result.Pledge.ToInt()
		res.Debt = result.Debt.ToInt()
		res.Entry = make([]RestrictingReleaseInfo, 0)
		for _, value := range result.Entry {
			res.Entry = append(res.Entry, RestrictingReleaseInfo{value.Height, value.Amount.ToInt()})
		}
		return res
	} else {
		return RestrictingResult{}
	}
}

func (n *Node) CallGetGovernParamValue(module, name string) (string, error) {
	xres := n.CallPPosContract(contract.GetGovernParamValue, module, name)
	if xres.Code == 0 {
		var result string
		result = xres.Ret.(string)
		return result, nil
	} else {
		return "", fmt.Errorf("query GovernParamValue fail,code:%v", xres.Code)
	}
}

//查询当前账户地址所委托的节点的NodeID和质押Id
func (n *Node) CallGetRelatedListByDelAddr(account common.Address) (staking.DelRelatedQueue, error) {
	xres := n.CallPPosContract(1103, account)
	var delRelatedQueue staking.DelRelatedQueue
	delRelatedQueue = xres.Ret.(staking.DelRelatedQueue)
	return delRelatedQueue, nil
}

func (n *Node) CallGetDelegateReward(account common.Address, nodes []discover.NodeID) (interface{}, error) {
	xres := n.CallPPosContract(5100, account, nodes)
	return xres.Ret, nil
}

func (n *Node) CallCandidateInfo(nodeID discover.NodeID) map[string]interface{} {
	xres := n.CallPPosContract(1105, nodeID)
	if xres.Code != 0 {
		return nil
	}
	var result map[string]interface{}
	result = xres.Ret.(map[string]interface{})
	return result
}

func (n *Node) GetVerifierList() []interface{} {
	xres := n.CallPPosContract(1100)
	if xres.Code != 0 {
		return nil
	}
	var result []interface{}
	result = xres.Ret.([]interface{})
	return result
}

func (n *Node) GetDelegateInfo(stakingBlockNum uint64, delAddr common.Address, nodeId discover.NodeID) map[string]interface{} {
	xres := n.CallPPosContract(contract.QueryDelegateInfo, stakingBlockNum, delAddr, nodeId)
	if xres.Code == 0 {
		var result map[string]interface{}
		result = xres.Ret.(map[string]interface{})
		return result
	} else {
		return nil
	}
}

func (n *Node) GetGetValidatorList() []interface{} {
	xres := n.CallPPosContract(1101)
	if xres.Code == 0 {
		var result []interface{}
		result = xres.Ret.([]interface{})
		return result
	} else {
		return nil
	}
}

//等待交易确认
func (n *Node) WaitTransactionByHash(txHash common.Hash) error {
	i := 0
	for {
		if i > 10 {
			return errors.New("wait to long")
		}
		_, isPending, err := n.client.TransactionByHash(n.ctx, txHash)
		if err == nil {
			if !isPending {
				break
			}
		} else {
			if err != ethereum.NotFound {
				return err
			}
		}
		time.Sleep(time.Second)
		i++
	}
	return nil
}

func (n *Node) GetXcomResult(txHash common.Hash) (*xcom.Result, error) {
	receipt, err := n.client.TransactionReceipt(n.ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("get TransactionReceipt fail:%v", err)
	}
	var logRes [][]byte
	if err := rlp.DecodeBytes(receipt.Logs[0].Data, &logRes); err != nil {
		return nil, fmt.Errorf("rlp decode fail:%v", err)
	}
	var res xcom.Result
	if err := json.Unmarshal(logRes[0], &res); err != nil {
		return nil, fmt.Errorf("json decode fail:%v", err)
	}
	return &res, nil
}
