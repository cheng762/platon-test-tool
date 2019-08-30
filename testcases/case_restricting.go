package testcases

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PlatONnetwork/PlatON-Go/common/vm"
	"github.com/PlatONnetwork/PlatON-Go/core/types"
	"github.com/PlatONnetwork/PlatON-Go/p2p/discover"
	"github.com/PlatONnetwork/PlatON-Go/x/restricting"
	"io/ioutil"
	"log"
	"math/big"
	"path"
)

type restrictCases struct {
	base       commonCases
	paramsPath string
	params     restrictingParams
	cases      map[string]func() error
}

type restrictingParams struct {
	CreatePlan       []restricting.RestrictingPlan `json:"create_plan"`
	CasePledgeReturn casePledgeReturnConfig        `json:"case_pledge_return"`
}

type casePledgeReturnConfig struct {
	Staking struct {
		BlsKey string `json:"bls_key"`
		NodeId string `json:"node_id"`
	} `json:"staking"`
}

func (r *restrictCases) Prepare() error {
	if err := r.base.Prepare(); err != nil {
		return err
	}
	r.paramsPath = path.Join(config.Dir, config.RestrictingConfigFile)
	r.parseConfigJson()
	r.registerfunc()
	return nil
}

func (r *restrictCases) registerfunc() {
	r.cases = make(map[string]func() error)
	r.registerTestCase("create_plan", r.caseCreatePlan)
	r.registerTestCase("pledgelock_return", r.casePledgeLockAndReturn)
}

func (r *restrictCases) Start() error {
	if err := r.caseCreatePlan(); err != nil {
		return r.base.SendError("restricting createPlan", err)
	}
	return nil
}

func (r *restrictCases) Exec(caseName string) error {
	f, ok := r.cases[caseName]
	if !ok {
		return fmt.Errorf("not found the case:%v", caseName)
	}
	if err := f(); err != nil {
		return err
	}
	return nil
}

func (r *restrictCases) End() error {
	if err := r.base.End(); err != nil {
		return err
	}
	if len(r.base.errors) != 0 {
		for _, value := range r.base.errors {
			r.base.SendError("restricting", value)
		}
		return errors.New("run restrictCases fail")
	}
	return nil
}

func (r *restrictCases) List() []string {
	r.registerfunc()
	var names []string
	for key, _ := range r.cases {
		names = append(names, key)
	}
	return names
}

func (r *restrictCases) registerTestCase(name string, f func() error) {
	r.cases[name] = f
}

func (r *restrictCases) parseConfigJson() {
	bytes, err := ioutil.ReadFile(r.paramsPath)
	if err != nil {
		panic(fmt.Errorf("parse restrictCases config file error,%s", err.Error()))
	}
	if err := json.Unmarshal(bytes, &r.params); err != nil {
		panic(fmt.Errorf("parse restrictCases config to json error,%s", err.Error()))
	}
}

func (r *restrictCases) caseCreatePlan() error {
	ctx := context.Background()
	from, _ := GetRandomAddr(r.base.adrs)
	plans := r.params.CreatePlan
	totalAmount := new(big.Int)
	for _, value := range plans {
		totalAmount.Add(totalAmount, value.Amount)
	}
	_, to := r.base.generateEmptyAccount()

	oldRestrictingContractAddr := r.base.GetBalance(ctx, vm.RestrictingContractAddr, nil)
	oldTo := r.base.GetBalance(ctx, to, nil)
	oldFrom := r.base.GetBalance(ctx, from, nil)
	txHash, err := r.base.CreateRestrictingPlanTransaction(ctx, from, to, plans)
	if err != nil {
		return err
	}
	if err := r.base.WaitTransactionByHash(ctx, txHash); err != nil {
		return fmt.Errorf("wait Transaction fail:%v", err)
	}

	//balance on RestrictingContractAddr
	newRestricting := r.base.GetBalance(ctx, vm.RestrictingContractAddr, nil)

	if new(big.Int).Sub(newRestricting, oldRestrictingContractAddr).Cmp(totalAmount) != 0 {
		return fmt.Errorf("RestrictingContractAddr balance is wrong,want %v,have %v", new(big.Int).Sub(newRestricting, oldRestrictingContractAddr), totalAmount)
	}

	tx, _, err := r.base.client.TransactionByHash(ctx, txHash)
	if err != nil {
		return err
	}

	totalGasAmountUsed := new(big.Int).Mul(tx.GasPrice(), new(big.Int).SetInt64(int64(tx.Gas())))
	tmp2 := new(big.Int).Add(totalAmount, totalGasAmountUsed)

	//balance on from
	newFrom := r.base.GetBalance(ctx, from, nil)
	if new(big.Int).Sub(oldFrom, newFrom).Cmp(tmp2) != 0 {
		return fmt.Errorf("from account %v  balance is wrong,want %v,have %v", from.String(), tmp2, new(big.Int).Sub(oldFrom, newFrom))
	}
	//balance on to
	newTo := r.base.GetBalance(ctx, to, nil)
	if newTo.Cmp(oldTo) != 0 {
		return fmt.Errorf("to account %v balance is wrong,want %v,have %v", to.String(), oldTo, newTo)
	}
	result := r.base.CallGetRestrictingInfo(ctx, to)
	log.Printf("plans: %+v", result)
	resAmount := new(big.Int)
	cmpfunc := func(block *types.Block, params ...interface{}) (bool, error) {
		height := params[0].(uint64)
		amount := params[1].(*big.Int)
		if block.Number().Uint64() > height {
			balance := r.base.GetBalance(ctx, to, nil)
			if balance.Cmp(amount) != 0 {
				return false, r.base.SendError("createPlan", fmt.Errorf("amount not comprare,want %v,have %v", amount, balance))
			}
			resBalance := r.base.GetBalance(ctx, vm.RestrictingContractAddr, nil)
			if new(big.Int).Sub(newRestricting, resBalance).Cmp(amount) != 0 {
				return false, r.base.SendError("createPlan", fmt.Errorf("RestrictingContractAddr amount not comprare,want %v,have %v", amount, new(big.Int).Sub(newRestricting, resBalance)))
			}
			log.Print("done cmp func")
			return true, nil
		}
		log.Printf("schedule:cal restricting release , current is %v,not arrive %v,wait for next schedule", block.Number(), height)
		return false, nil
	}
	for _, val := range result.Entry {
		tmp := new(big.Int).Add(resAmount, val.Amount.ToInt())
		resAmount.Add(resAmount, val.Amount.ToInt())
		r.base.addJobs(cmpfunc, val.Height, tmp)
	}
	return nil
}

func (r *restrictCases) casePledgeLockAndReturn() error {
	ctx := context.Background()
	VersionValue, err := r.base.CallProgramVersion(ctx)
	if err != nil {
		return err
	}
	from1 := r.base.adrs[1]
	var input stakingInput
	input.BlsPubKey = r.params.CasePledgeReturn.Staking.BlsKey
	input.Amount, _ = new(big.Int).SetString("10000000000000000000000000", 10)
	input.Typ = 0
	input.BenefitAddress = r.base.adrs[1]
	id, err := discover.HexID(r.params.CasePledgeReturn.Staking.NodeId)
	if err != nil {
		return err
	}
	input.NodeId = id

	log.Print("begin create staking")
	txhash2, err := r.base.CreateStakingTransaction(ctx, from1, input, VersionValue)
	if err != nil {
		return fmt.Errorf("createStakingTransaction fail:%v", err)
	}
	if err := r.base.WaitTransactionByHash(ctx, txhash2); err != nil {
		return fmt.Errorf("wait Transaction2 %v fail:%v", txhash2, err)
	}
	log.Print("end create staking", txhash2.String())

	log.Print("begin create restricting plans")
	from2 := r.base.adrs[2]
	plans := make([]restricting.RestrictingPlan, 0)

	amount, _ := new(big.Int).SetString("10000000000000000000000000", 10)
	plans = append(plans, restricting.RestrictingPlan{1, amount})
	to := r.base.adrs[3]
	txHash, err := r.base.CreateRestrictingPlanTransaction(ctx, from2, to, plans)
	if err != nil {
		return fmt.Errorf("CreateRestrictingPlanTransaction fail:%v", err)
	}
	if err := r.base.WaitTransactionByHash(ctx, txHash); err != nil {
		return fmt.Errorf("wait Transaction %v fail:%v", txHash, err)
	}

	amount2, _ := new(big.Int).SetString("10000000000000000000000000", 10)
	log.Print("begin delegateTransaction", amount2)

	txhash4, err := r.base.DelegateTransaction(ctx, to, id, 1, amount2)
	if err != nil {
		return fmt.Errorf("delegateTransaction fail:%v", err)
	}
	if err := r.base.WaitTransactionByHash(ctx, txhash4); err != nil {
		return fmt.Errorf("wait Transaction %v fail:%v", txhash4, err)
	}

	log.Print("end delegateTransaction", txhash4.String())

	res := r.base.CallGetRestrictingInfo(ctx, to)
	log.Printf("RestrictingInfo:%+v", res)
	quene, err := r.base.CallGetRelatedListByDelAddr(ctx, to)
	if err != nil {
		return fmt.Errorf("getRelatedListByDelAddr fail:%v", err)
	}
	log.Print("getRelatedListByDelAddr", quene)
	tmpfunc := func(block *types.Block, params ...interface{}) (bool, error) {
		hight := params[0].(uint64)
		log.Printf("wait %v,now %v,condition %v", hight, block.Number().Uint64(), block.Number().Uint64() > hight)
		if block.Number().Uint64() > hight {
			balance := params[1].(*big.Int)
			txhash, err := r.base.WithdrewDelegateTransaction(ctx, quene[0].StakingBlockNum, id, to, balance)
			if err != nil {
				return false, err
			}
			if err := r.base.WaitTransactionByHash(ctx, txhash); err != nil {
				return false, err
			}
			result := r.base.CallGetRestrictingInfo(ctx, to)
			log.Printf("plans: %+v", result)
			return true, nil
		}
		return false, nil
	}
	amount3, _ := new(big.Int).SetString("5000000000000000000000000", 10)
	r.base.addJobs(tmpfunc, res.Entry[0].Height, amount3)
	r.base.addJobs(tmpfunc, res.Entry[0].Height+300, amount3)
	return nil
}
