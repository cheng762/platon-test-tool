package testcases

import (
	"context"
	"errors"
	"github.com/PlatONnetwork/PlatON-Go/params"
	"github.com/PlatONnetwork/PlatON-Go/x/gov"
	"github.com/PlatONnetwork/PlatON-Go/x/plugin"
	"github.com/PlatONnetwork/platon-test-tool/Factory"
	"log"
	"math/big"
	"path"
)

type stakingCases struct {
	commonCases
	paramsPath string
}

func (r *stakingCases) Prepare() error {
	if err := r.commonCases.Prepare(); err != nil {
		return err
	}
	r.paramsPath = path.Join(config.Dir, config.StakingConfigFile)
	return nil
}

func (r *stakingCases) Start() error {
	return nil
}

func (r *stakingCases) Exec(caseName string) error {
	return r.commonCases.exec(caseName, r)
}

func (r *stakingCases) End() error {
	if err := r.commonCases.End(); err != nil {
		return err
	}
	if len(r.errors) != 0 {
		for _, value := range r.errors {
			r.SendError("restricting", value)
		}
		return errors.New("run restrictCases fail")
	}
	return nil
}

func (r *stakingCases) List() []string {
	return r.list(r)
}

//this test is test in one block,staking,delegate,unstaking,staking,delegate, the two delegate will merge
func (r *stakingCases) CaseMergeDelegate() error {
	ctx := context.Background()
	//查询质押门槛

	node := r.TxManager.GetNode()

	node.SendTraction()

	threshold, err := r.CallGetGovernParamValue(ctx, gov.ModuleStaking, gov.KeyStakeThreshold)
	if err != nil {
		return err
	}
	log.Printf("获取质押门槛%v", threshold)

	OperatingThreshold, err := r.CallGetGovernParamValue(ctx, gov.ModuleStaking, gov.KeyOperatingThreshold)
	if err != nil {
		return err
	}
	log.Printf("获取委托门槛%v", OperatingThreshold)

	stakingAmount, _ := new(big.Int).SetString(threshold, 10)

	delgateAmount, _ := new(big.Int).SetString(OperatingThreshold, 10)

	config := Factory.LoadStakingConfig(r.paramsPath)
	input, err := Factory.BuildStaking(config, stakingAmount)
	if err != nil {
		return err
	}

	delfaultGasPrice := params.GVon

	stakingAccount, _ := AccountPool.Get().(*PriAccount)
	log.Printf("begin create staking,account:%v,info:%+v", stakingAccount.Address.Hex(), input)
	stakingAccount.gasPrice = new(big.Int).SetUint64(uint64(delfaultGasPrice) + 1000)
	stakingTransaction, err := r.CreateStakingTransaction(ctx, stakingAccount, *input)
	if err != nil {
		return r.SendError("CaseMergeDelegate", err)
	}
	log.Printf("第一次质押交易发起成功，交易hash:%v", stakingTransaction.Hex())

	delAccount, _ := AccountPool.Get().(*PriAccount)

	log.Printf("第一次发起委托交易,account:%v", delAccount.Address.Hex())
	delAccount.gasPrice = new(big.Int).SetUint64(uint64(delfaultGasPrice) + 800)
	deltx, err := r.DelegateTransaction(ctx, delAccount, input.NodeId, plugin.FreeVon, delgateAmount)
	if err != nil {
		return r.SendError("CaseMergeDelegate", err)
	}
	log.Printf("第一次委托交易发起成功，交易hash:%v", deltx.Hex())

	log.Printf("第一次发起解除质押交易")
	stakingAccount.gasPrice = new(big.Int).SetUint64(uint64(delfaultGasPrice) + 600)
	txWithdrewStaking, err := r.WithdrewStaking(ctx, stakingAccount, input.NodeId)
	if err != nil {
		return r.SendError("CaseMergeDelegate", err)
	}
	log.Printf("第一次发起解除质押交易发起成功，交易hash:%v", txWithdrewStaking.Hex())

	log.Printf("第二次质押,account:%v,info:%+v", stakingAccount.Address.Hex(), input)
	stakingAccount.gasPrice = new(big.Int).SetUint64(uint64(delfaultGasPrice) + 400)
	txstaking2, err := r.CreateStakingTransaction(ctx, stakingAccount, *input)
	if err != nil {
		return r.SendError("CaseMergeDelegate", err)
	}
	log.Printf("第二次质押交易发起成功，交易hash:%v", txstaking2.Hex())

	log.Printf("第二次发起委托交易,account:%v", delAccount.Address.Hex())
	delAccount.gasPrice = new(big.Int).SetUint64(uint64(delfaultGasPrice) + 200)
	txdel2, err := r.DelegateTransaction(ctx, delAccount, input.NodeId, plugin.FreeVon, delgateAmount)
	if err != nil {
		return r.SendError("CaseMergeDelegate", err)
	}
	log.Printf("第二次委托交易发起成功，交易hash:%v", txdel2.Hex())

	return nil
}
