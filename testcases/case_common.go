package testcases

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	ethereum "github.com/PlatONnetwork/PlatON-Go"
	"github.com/PlatONnetwork/PlatON-Go/common"
	"github.com/PlatONnetwork/PlatON-Go/common/vm"
	"github.com/PlatONnetwork/PlatON-Go/core/types"
	"github.com/PlatONnetwork/PlatON-Go/crypto"
	"github.com/PlatONnetwork/PlatON-Go/ethclient"
	"github.com/PlatONnetwork/PlatON-Go/node"
	"github.com/PlatONnetwork/PlatON-Go/p2p/discover"
	"github.com/PlatONnetwork/PlatON-Go/params"
	"github.com/PlatONnetwork/PlatON-Go/rlp"
	"github.com/PlatONnetwork/PlatON-Go/x/restricting"
	"github.com/PlatONnetwork/PlatON-Go/x/staking"
	"github.com/PlatONnetwork/PlatON-Go/x/xcom"
	"github.com/PlatONnetwork/platon-test-tool/Dto"
	common2 "github.com/PlatONnetwork/platon-test-tool/common"

	contract "github.com/PlatONnetwork/PlatON-Go/core/vm"

	"github.com/robfig/cron"
	"log"
	"math/big"
	"reflect"
	"strings"
	"time"
)

const PrefixCase = "Case"

type commonCases struct {
	TxManager common2.TxManager
	//adrs   []common.Address
	corn   *cron.Cron
	jobs   []*job
	errors []error
	errch  chan error
	donech chan struct{}
}

type job struct {
	handle func(block *types.Block, params ...interface{}) (bool, error)
	params []interface{}
	done   bool
	runing bool
	desc   string
}

func (j *job) run(block *types.Block) (bool, error) {
	return j.handle(block, j.params...)
}

func (c *commonCases) list(m caseTest) []string {
	var names []string
	object := reflect.TypeOf(m)
	for i := 0; i < object.NumMethod(); i++ {
		method := object.Method(i)
		if strings.HasPrefix(method.Name, PrefixCase) {
			names = append(names, strings.TrimPrefix(method.Name, "Case"))
		}
	}
	return names
}

func (c *commonCases) exec(caseName string, m caseTest) error {
	methodNames := PrefixCase + caseName
	val := reflect.ValueOf(m).MethodByName(methodNames).Call([]reflect.Value{})
	if val[0].IsNil() {
		return nil
	}
	err := val[0].Interface().(error)
	return err
}

//初始化case
func (c *commonCases) Prepare() error {
	client, err := ethclient.Dial(config.Url)
	if err != nil {
		return err
	}
	c.client = client
	//	c.adrs = GetAllAddress()
	c.corn = cron.New()
	if _, err := c.corn.AddFunc("@every 1s", c.schedule); err != nil {
		return err
	}
	c.donech = make(chan struct{}, 0)
	c.errch = make(chan error, 0)
	c.errors = make([]error, 0)
	return nil
}

func (c *commonCases) Start() error {
	return nil
}

func (c *commonCases) End() error {
	c.corn.Start()
	for {
		select {
		case err := <-c.errch:
			c.errors = append(c.errors, err)
		case <-c.donech:
			c.corn.Stop()
			close(c.errch)
			return nil
		}
	}
}

func (c *commonCases) GetNonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	return c.client.NonceAt(ctx, account, blockNumber)
}

func (c *commonCases) encodePPOS(funcType uint16, params ...interface{}) []byte {
	par := buildParams(funcType, params...)
	buf := new(bytes.Buffer)
	err := rlp.Encode(buf, par)
	if err != nil {
		panic(fmt.Errorf("encode rlp data fail: %v", err))
	}
	return buf.Bytes()
}

func (c *commonCases) GetXcomResult(ctx context.Context, txHash common.Hash) (*xcom.Result, error) {
	receipt, err := c.client.TransactionReceipt(ctx, txHash)
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

//等待交易确认
func (c *commonCases) WaitTransactionByHash(ctx context.Context, txHash common.Hash) error {
	i := 0
	for {
		if i > 10 {
			return errors.New("wait to long")
		}
		_, isPending, err := c.client.TransactionByHash(ctx, txHash)
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

//获取ProgramVersion
func (c *commonCases) GetSchnorrNIZKProve(ctx context.Context) (string, error) {
	prove, err := c.client.GetSchnorrNIZKProve(ctx)
	if err != nil {
		return "", err
	}
	return prove, nil
}

//获取ProgramVersion
func (c *commonCases) CallProgramVersion(ctx context.Context) (*params.ProgramVersion, error) {
	pg, err := c.client.GetProgramVersion(ctx)
	if err != nil {
		return nil, err
	}
	return pg, nil
}

func (c *commonCases) WithdrawDelegateReward(ctx context.Context, from *PriAccount) (common.Hash, error) {
	return c.SendPPosTraction(ctx, from, contract.TxWithdrawDelegateReward)
}

func (c *commonCases) DeclareVersion(ctx context.Context, from *PriAccount, nodePrikey *ecdsa.PrivateKey, activeNode discover.NodeID, programVersion uint32) (common.Hash, error) {
	handle := node.GetCryptoHandler()
	handle.SetPrivateKey(nodePrikey)
	versionSign := common.VersionSign{}
	versionSign.SetBytes(handle.MustSign(programVersion))
	return c.SendPPosTraction(ctx, from, contract.Declare, activeNode, programVersion, versionSign)
}

//创建委托
func (c *commonCases) DelegateTransaction(ctx context.Context, account *PriAccount, nodeID discover.NodeID, typ uint16, amount *big.Int) (common.Hash, error) {
	return c.SendPPosTraction(ctx, account, contract.TxDelegate, typ, nodeID, amount)
}

//减持/撤销委托
func (c *commonCases) WithdrewDelegateTransaction(ctx context.Context, stakingBlockNum uint64, nodeID discover.NodeID, account *PriAccount, amount *big.Int) (common.Hash, error) {
	return c.SendPPosTraction(ctx, account, contract.TxWithdrewDelegate, stakingBlockNum, nodeID, amount)
}

func (c *commonCases) SendPPosTraction(ctx context.Context, account *PriAccount, funcType uint16, params ...interface{}) (common.Hash, error) {
	send := c.encodePPOS(funcType, params...)
	toadd := funcTypeToContractAddress(funcType)
	txhash, err := SendRawTransaction(ctx, c.client, account, *toadd, "0", send)
	if err != nil {
		return common.ZeroHash, err
	}
	return txhash, nil
}

//查询当前账户地址所委托的节点的NodeID和质押Id
func (c *commonCases) CallGetRelatedListByDelAddr(ctx context.Context, account common.Address) (staking.DelRelatedQueue, error) {
	xres := c.CallPPosContract(ctx, 1103, account)
	var delRelatedQueue staking.DelRelatedQueue
	delRelatedQueue = xres.Ret.(staking.DelRelatedQueue)
	return delRelatedQueue, nil
}

func (c *commonCases) CallGetDelegateReward(ctx context.Context, account common.Address, nodes []discover.NodeID) (interface{}, error) {
	xres := c.CallPPosContract(ctx, 5100, account, nodes)
	return xres.Ret, nil
}

// for plugin test
type restrictingResult struct {
	Balance *big.Int                 `json:"balance"`
	Debt    *big.Int                 `json:"debt"`
	Entry   []restrictingReleaseInfo `json:"plans"`
	Pledge  *big.Int                 `json:"Pledge"`
}

// for plugin test
type restrictingReleaseInfo struct {
	Height uint64   `json:"blockNumber"` // blockNumber representation of the block number at the released epoch
	Amount *big.Int `json:"amount"`      // amount representation of the released amount
}

func (c *commonCases) CallCandidateInfo(ctx context.Context, nodeID discover.NodeID) map[string]interface{} {
	xres := c.CallPPosContract(ctx, 1105, nodeID)
	if xres.Code != 0 {
		return nil
	}
	var result map[string]interface{}
	result = xres.Ret.(map[string]interface{})
	return result
}

func (c *commonCases) GetVerifierList(ctx context.Context) []interface{} {
	xres := c.CallPPosContract(ctx, 1100)
	if xres.Code != 0 {
		return nil
	}
	var result []interface{}
	result = xres.Ret.([]interface{})
	return result
}

func (c *commonCases) GetDelegateInfo(ctx context.Context, stakingBlockNum uint64, delAddr common.Address, nodeId discover.NodeID) map[string]interface{} {
	xres := c.CallPPosContract(ctx, contract.QueryDelegateInfo, stakingBlockNum, delAddr, nodeId)
	if xres.Code == 0 {
		var result map[string]interface{}
		result = xres.Ret.(map[string]interface{})
		return result
	} else {
		return nil
	}
}

func (c *commonCases) GetGetValidatorList(ctx context.Context) []interface{} {
	xres := c.CallPPosContract(ctx, 1101)
	if xres.Code == 0 {
		var result []interface{}
		result = xres.Ret.([]interface{})
		return result
	} else {
		return nil
	}
}

//获取锁仓信息
func (c *commonCases) CallGetRestrictingInfo(ctx context.Context, account common.Address) restrictingResult {
	xres := c.CallPPosContract(ctx, 4100, account)
	if xres.Code == 0 {
		var result restricting.Result
		result = xres.Ret.(restricting.Result)
		var res restrictingResult
		res.Balance = result.Balance.ToInt()
		res.Pledge = result.Pledge.ToInt()
		res.Debt = result.Debt.ToInt()
		res.Entry = make([]restrictingReleaseInfo, 0)
		for _, value := range result.Entry {
			res.Entry = append(res.Entry, restrictingReleaseInfo{value.Height, value.Amount.ToInt()})
		}
		return res
	} else {
		return restrictingResult{}
	}
}

func (c *commonCases) CallGetGovernParamValue(ctx context.Context, module, name string) (string, error) {
	xres := c.CallPPosContract(ctx, contract.GetGovernParamValue, module, name)
	if xres.Code == 0 {
		var result string
		result = xres.Ret.(string)
		return result, nil
	} else {
		return "", fmt.Errorf("query GovernParamValue fail,code:%v", xres.Code)
	}
}

func (c *commonCases) CallPPosContract(ctx context.Context, funcType uint16, params ...interface{}) xcom.Result {
	send := c.encodePPOS(funcType, params...)
	var msg ethereum.CallMsg
	msg.Data = send
	msg.To = funcTypeToContractAddress(funcType)
	res, err := c.client.CallContract(ctx, msg, nil)
	if err != nil {
		panic(err)
	}
	var xres xcom.Result
	if err := json.Unmarshal(res, &xres); err != nil {
		panic(err)
	}
	return xres
}

//获取账户金额
func (c *commonCases) GetBalance(ctx context.Context, address common.Address, blockNumber *big.Int) *big.Int {
	balance, err := c.client.BalanceAt(ctx, address, blockNumber)
	if err != nil {
		panic(err)
	}
	return balance
}

func (c *commonCases) SendError(caseName string, err error) error {
	log.Printf("[fail]test case %v fail: %v ", caseName, err)
	return fmt.Errorf("test case %v fail: %v ", caseName, err)
}

//添加需等待的测试，每隔1s将当前块高传入f中并执行
// f为需要执行的函数，当执行成功返回true，nil，当未达到执行条件返回false，nil，当执行失败返回false，error
// params为f所需要的参数
func (c *commonCases) addJobs(desc string, f func(block *types.Block, params ...interface{}) (bool, error), params ...interface{}) {
	var j *job
	j = new(job)
	j.params = params
	j.handle = f
	j.desc = desc
	c.jobs = append(c.jobs, j)
}

//生成一个新的空账户，私钥和地址
func (c *commonCases) generateEmptyAccount() (*ecdsa.PrivateKey, common.Address) {
	privateKey, _ := crypto.GenerateKey()
	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	var pri PriAccount
	pri.Priv = privateKey
	pri.Address = address
	allAccounts[address] = &pri
	return privateKey, address
}

func (c *commonCases) schedule() {
	block, err := c.client.BlockByNumber(context.Background(), nil)
	if err != nil {
		c.errch <- fmt.Errorf("job error: %v", err)
		c.donech <- struct{}{}
		close(c.donech)
		return
	}
	if block.Number().Uint64()%100 == 0 {
		log.Printf("schedule working,current block num is %v", block.Number())
	}
	for i := 0; i < len(c.jobs); {
		if c.jobs[i].done {
			log.Printf("[job] %v done", c.jobs[i].desc)
			c.jobs = append(c.jobs[:i], c.jobs[i+1:]...)
		} else {
			if !c.jobs[i].runing {
				go func(j *job) {
					defer func() {
						if r := recover(); r != nil {
							log.Printf("Recovered in :%v", r)
							c.errch <- fmt.Errorf("job %v painc", j.desc)
							j.done = true
						}
					}()
					j.runing = true
					ok, err := j.run(block)
					if err != nil {
						c.errch <- fmt.Errorf("job %v fail:%v", c.jobs[i].desc, err)
						j.done = true
						return
					}
					if ok {
						j.done = true
					}
					j.runing = false
				}(c.jobs[i])
			}
			i++
		}
	}
	if len(c.jobs) == 0 {
		c.donech <- struct{}{}
		close(c.donech)
	}
}

func buildParams(funcType uint16, params ...interface{}) [][]byte {
	var res [][]byte
	res = make([][]byte, 0)
	fnType, _ := rlp.EncodeToBytes(funcType)
	res = append(res, fnType)
	for _, param := range params {
		val, err := rlp.EncodeToBytes(param)
		if err != nil {
			panic(err)
		}
		res = append(res, val)
	}
	return res
}

func funcTypeToContractAddress(funcType uint16) *common.Address {
	toadd := common.ZeroAddr
	switch {
	case 0 < funcType && funcType < 2000:
		toadd = vm.StakingContractAddr
	case funcType >= 2000 && funcType < 3000:
		toadd = vm.GovContractAddr
	case funcType >= 3000 && funcType < 4000:
		toadd = vm.SlashingContractAddr
	case funcType >= 4000 && funcType < 5000:
		toadd = vm.RewardManagerPoolAddr
	case funcType >= 5000 && funcType < 6000:
		toadd = vm.DelegateRewardPoolAddr
	}
	return &toadd
}
