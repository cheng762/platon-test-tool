package testcases

import (
	"errors"
	"log"
	"math/big"

	"github.com/PlatONnetwork/platon-test-tool/common"
	"github.com/PlatONnetwork/platon-test-tool/config"

	"github.com/PlatONnetwork/platon-test-tool/Factory"

	"github.com/PlatONnetwork/PlatON-Go/params"
	"github.com/PlatONnetwork/PlatON-Go/x/gov"
	"github.com/PlatONnetwork/PlatON-Go/x/plugin"
)

type stakingCases struct {
	*commonCases
	config config.StakingCaseConfig
}

func (r *stakingCases) Prepare() error {
	r.commonCases = NewCommonCases(TxManager.GetRandomNode())
	path := TxManager.Config.GetCaseConfigPath("staking")
	return common.LoadConfig(path, &r.config)
}

func (r *stakingCases) Start() error {
	return nil
}

func (r *stakingCases) Exec(caseName string) error {
	return execCase(caseName, r)
}

func (r *stakingCases) End() error {
	if err := r.commonCases.End(); err != nil {
		return err
	}
	if len(r.errors) != 0 {
		for _, value := range r.errors {
			SendError("restricting", value)
		}
		return errors.New("run restrictCases fail")
	}
	return nil
}

func (r *stakingCases) List() []string {
	return listCase(r)
}

//this test is test in one block,staking,delegate,unstaking,staking,delegate, the two delegate will merge
func (r *stakingCases) CaseMergeDelegate() error {
	//查询质押门槛

	node := TxManager.GetNode(r.config.MergeDelegate.NodeUrl)

	threshold, err := node.CallGetGovernParamValue(gov.ModuleStaking, gov.KeyStakeThreshold)
	if err != nil {
		return err
	}
	log.Printf("获取质押门槛%v", threshold)

	OperatingThreshold, err := node.CallGetGovernParamValue(gov.ModuleStaking, gov.KeyOperatingThreshold)
	if err != nil {
		return err
	}
	log.Printf("获取委托门槛%v", OperatingThreshold)

	stakingAmount, _ := new(big.Int).SetString(threshold, 10)

	delgateAmount, _ := new(big.Int).SetString(OperatingThreshold, 10)

	input, err := Factory.BuildStaking(r.config.MergeDelegate, stakingAmount)
	if err != nil {
		return err
	}

	delfaultGasPrice := params.GVon

	stakingAccount := TxManager.GetAccount()
	log.Printf("begin create staking,account:%v,info:%+v", stakingAccount.Address.Hex(), input)
	stakingAccount.GasPrice = new(big.Int).SetUint64(uint64(delfaultGasPrice) + 1000)

	stakingTransaction, err := stakingAccount.CreateStakingTransaction(node, *input)
	if err != nil {
		return err
	}
	node.AddRawTraction(stakingTransaction)

	log.Printf("第一次质押交易发起成功，交易hash:%v", stakingTransaction.Hash().Hex())

	delAccount := TxManager.GetAccount()

	log.Printf("第一次发起委托交易,account:%v", delAccount.Address.Hex())
	delAccount.GasPrice = new(big.Int).SetUint64(uint64(delfaultGasPrice) + 800)

	deltx, err := delAccount.DelegateTransaction(node, input.NodeId, plugin.FreeVon, delgateAmount)
	if err != nil {
		return err
	}

	log.Printf("第一次委托交易发起成功，交易hash:%v", deltx.Hash().Hex())

	node.AddRawTraction(deltx)

	log.Printf("第一次发起解除质押交易")
	stakingAccount.GasPrice = new(big.Int).SetUint64(uint64(delfaultGasPrice) + 600)
	txWithdrewStaking, err := stakingAccount.WithdrewStaking(node, input.NodeId)
	if err != nil {
		return err
	}
	node.AddRawTraction(txWithdrewStaking)

	log.Printf("第一次发起解除质押交易发起成功，交易hash:%v", txWithdrewStaking.Hash().Hex())

	log.Printf("第二次质押,account:%v,info:%+v", stakingAccount.Address.Hex(), input)
	stakingAccount.GasPrice = new(big.Int).SetUint64(uint64(delfaultGasPrice))
	txstaking2, err := stakingAccount.CreateStakingTransaction(node, *input)
	if err != nil {
		return err
	}
	node.AddRawTraction(txstaking2)

	log.Printf("第二次质押交易发起成功，交易hash:%v", txstaking2.Hash().Hex())

	log.Printf("第二次发起委托交易,account:%v", delAccount.Address.Hex())
	delAccount.GasPrice = new(big.Int).SetUint64(uint64(delfaultGasPrice) + 200)

	txdel2, err := delAccount.DelegateTransaction(node, input.NodeId, plugin.FreeVon, delgateAmount)
	if err != nil {
		return err
	}
	node.AddRawTraction(txdel2)
	log.Printf("第二次委托交易发起成功，交易hash:%v", txdel2.Hash().Hex())

	if err := node.SendAllTraction(); err != nil {
		return err
	}

	return nil
}
