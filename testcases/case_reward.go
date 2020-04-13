package testcases

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PlatONnetwork/PlatON-Go/common"
	"github.com/PlatONnetwork/PlatON-Go/common/hexutil"
	"github.com/PlatONnetwork/PlatON-Go/common/vm"
	"github.com/PlatONnetwork/PlatON-Go/core/types"
	"github.com/PlatONnetwork/PlatON-Go/crypto"
	"github.com/PlatONnetwork/PlatON-Go/crypto/bls"
	"github.com/PlatONnetwork/PlatON-Go/node"
	"github.com/PlatONnetwork/PlatON-Go/p2p/discover"
	"github.com/PlatONnetwork/PlatON-Go/params"
	"github.com/PlatONnetwork/PlatON-Go/rlp"
	"github.com/PlatONnetwork/PlatON-Go/x/reward"
	"github.com/PlatONnetwork/PlatON-Go/x/xcom"
	"github.com/PlatONnetwork/platon-test-tool/Dto"
	"io/ioutil"
	"log"
	"math/big"
	"path"
	"time"
)

type rewardCases struct {
	commonCases
	paramsPath string
	params     caseRewardConfig
}

type caseRewardConfig struct {
	DelegateReward struct {
		BlsKey          string `json:"bls_key"`
		NodeId          string `json:"node_id"`
		DelegateAccount string `json:"delegate_account"`
		StakingAccount  string `json:"staking_account"`
		StakingNum      uint64 `json:"staking_num"`
	} `json:"delegate_reward"`
}

func (r *rewardCases) parseConfigJson() {
	bytes, err := ioutil.ReadFile(r.paramsPath)
	if err != nil {
		panic(fmt.Errorf("parse restrictCases config file error,%s", err.Error()))
	}
	if err := json.Unmarshal(bytes, &r.params); err != nil {
		panic(fmt.Errorf("parse restrictCases config to json error,%s", err.Error()))
	}
}

func (r *rewardCases) saveConfigJson() {
	v, err := json.Marshal(r.params)
	if err != nil {
		panic(fmt.Errorf("parse restrictCases config to json error,%s", err.Error()))
	}
	err = ioutil.WriteFile(r.paramsPath, v, 0644)
	if err != nil {
		panic(fmt.Errorf("write to addr.json error%s \n", err.Error()))
	}
}

func (r *rewardCases) Prepare() error {
	if err := r.commonCases.Prepare(); err != nil {
		return err
	}
	r.paramsPath = path.Join(config.Dir, config.RewardConfigFile)
	r.parseConfigJson()
	return nil
}

func (r *rewardCases) Start() error {

	return nil
}

func (r *rewardCases) Exec(caseName string) error {
	return r.commonCases.exec(caseName, r)
}

func (r *rewardCases) End() error {
	r.saveConfigJson()
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

func (r *rewardCases) List() []string {
	return r.list(r)
}

func (r *rewardCases) CaseRewardPoolBalance() error {
	ctx := context.Background()
	totalRewardBalance := r.GetBalance(ctx, vm.RewardManagerPoolAddr, big.NewInt(0))
	totalCommunityDeveloper := r.GetBalance(ctx, xcom.CDFAccount(), big.NewInt(0))
	totalPlatONFoundation := r.GetBalance(ctx, xcom.PlatONFundAccount(), big.NewInt(0))
	res := r.CallGetRestrictingInfo(ctx, vm.RewardManagerPoolAddr)
	log.Printf("Restricting plan:%+v", res)

	f := func(block *types.Block, params ...interface{}) (bool, error) {
		year := params[0].(int)
		if block.Number().Int64() >= int64(year*1600) {
			res2 := r.CallGetRestrictingInfo(ctx, vm.RewardManagerPoolAddr)
			log.Printf("block:%v Restricting plan:%+v", block, res2)
			Reward := params[1].(*big.Int)
			Foundation := params[2].(*big.Int)
			Community := params[3].(*big.Int)
			CommunityDeveloperBalance := r.GetBalance(ctx, xcom.CDFAccount(), nil)
			PlatONFoundationBalance := r.GetBalance(ctx, xcom.PlatONFundAccount(), nil)
			rewardBalance := r.GetBalance(ctx, vm.RewardManagerPoolAddr, nil)
			log.Printf("reward: want %v have %v", Reward, rewardBalance)
			log.Printf("CommunityDeveloperBalance: want %v have %v", Community, CommunityDeveloperBalance)
			log.Printf("PlatONFoundationBalance: want %v have %v", Foundation, PlatONFoundationBalance)
			return true, nil
		}
		return false, nil
	}
	totalAmount, _ := new(big.Int).SetString("10250000000000000000000000000", 10)
	for i := 0; i < 12; i++ {
		increse := new(big.Int).Div(totalAmount, big.NewInt(40))
		tmp := new(big.Int).Mul(increse, big.NewInt(80))
		giveReward := new(big.Int).Div(tmp, big.NewInt(100))
		if i < len(res.Entry) {
			totalRewardBalance.Add(totalRewardBalance, res.Entry[i].Amount)
		}
		if i < 10 {
			giveCommunityDeveloper := new(big.Int).Sub(increse, giveReward)
			totalRewardBalance.Add(totalRewardBalance, giveReward)
			totalCommunityDeveloper.Add(totalCommunityDeveloper, giveCommunityDeveloper)
		} else {
			tmp3 := new(big.Int).Sub(increse, giveReward)
			giveCommunityDeveloper := new(big.Int).Div(tmp3, big.NewInt(2))
			givePlatONFoundation := new(big.Int).Div(tmp3, big.NewInt(2))
			totalCommunityDeveloper.Add(totalCommunityDeveloper, giveCommunityDeveloper)
			totalRewardBalance.Add(totalRewardBalance, giveReward)
			totalPlatONFoundation.Add(totalPlatONFoundation, givePlatONFoundation)
		}
		r.addJobs(fmt.Sprintf("cal year %d reward", i), f, i+1, new(big.Int).Set(totalRewardBalance), new(big.Int).Set(totalPlatONFoundation), new(big.Int).Set(totalCommunityDeveloper))

		totalAmount.Add(totalAmount, increse)
	}
	return nil
}

func (r *rewardCases) CaseIncreseBalance() error {

	totalCommunityDeveloper, _ := new(big.Int).SetString("327311981000000000000000000", 10)
	PlatONFoundationAddress := new(big.Int)

	var (
		zeroEpoch  = new(big.Int).Mul(big.NewInt(62215742), big.NewInt(1e18))
		oneEpoch   = new(big.Int).Mul(big.NewInt(55965742), big.NewInt(1e18))
		twoEpoch   = new(big.Int).Mul(big.NewInt(49559492), big.NewInt(1e18))
		threeEpoch = new(big.Int).Mul(big.NewInt(42993086), big.NewInt(1e18))
		fourEpoch  = new(big.Int).Mul(big.NewInt(36262520), big.NewInt(1e18))
		fiveEpoch  = new(big.Int).Mul(big.NewInt(29363689), big.NewInt(1e18))
		sixEpoch   = new(big.Int).Mul(big.NewInt(22292388), big.NewInt(1e18))
		sevenEpoch = new(big.Int).Mul(big.NewInt(15044304), big.NewInt(1e18))
		eightEpoch = new(big.Int).Mul(big.NewInt(7615018), big.NewInt(1e18))
	)
	var plans []*big.Int = []*big.Int{zeroEpoch, oneEpoch, twoEpoch, threeEpoch, fourEpoch, fiveEpoch, sixEpoch, sevenEpoch, eightEpoch}
	rewardBalance, _ := new(big.Int).SetString("200000000000000000000000000", 10)

	rewardBalance.Add(rewardBalance, zeroEpoch)
	log.Print(rewardBalance)
	for _, value := range plans {
		totalCommunityDeveloper.Sub(totalCommunityDeveloper, value)
	}

	totalAmount, _ := new(big.Int).SetString("10250000000000000000000000000", 10)

	stakingAmoount, _ := new(big.Int).SetString("1500000000000000000000000", 10)
	stakingAmoount.Mul(stakingAmoount, big.NewInt(4))

	totalCommunityDeveloper.Sub(totalCommunityDeveloper, stakingAmoount)

	log.Printf("totalCommunityDeveloper value:%v", totalCommunityDeveloper)

	for i := 1; i < 15; i++ {
		increse := new(big.Int).Mul(totalAmount, big.NewInt(25))
		increse.Div(increse, big.NewInt(1000))
		tmp := new(big.Int).Mul(increse, big.NewInt(80))
		log.Print(tmp)
		giveReward := new(big.Int).Div(tmp, big.NewInt(100))
		if i < 9 {
			rewardBalance.Add(rewardBalance, plans[i])
		}
		if i < 9 {
			giveCommunityDeveloper := new(big.Int).Sub(increse, giveReward)
			totalCommunityDeveloper.Add(totalCommunityDeveloper, giveCommunityDeveloper)
		} else {
			giveOther := new(big.Int).Sub(increse, giveReward)
			tmp := new(big.Int).Mul(giveOther, big.NewInt(5))
			tmp2 := new(big.Int).Div(tmp, big.NewInt(10))
			PlatONFoundationAddress.Add(PlatONFoundationAddress, tmp2)
			totalCommunityDeveloper.Add(totalCommunityDeveloper, tmp2)
		}
		rewardBalance.Add(rewardBalance, giveReward)
		totalAmount.Add(totalAmount, increse)
		log.Printf("year %v increase:%v totalAmount %v giveReward:%v  rewardBalance %v CommunityDeveloper %v PlatONFoundation%v", i, increse, totalAmount, giveReward, rewardBalance, totalCommunityDeveloper, PlatONFoundationAddress)
	}
	return nil
}

func (r *rewardCases) CaseCal() error {
	b, _ := new(big.Int).SetString("262215742486916500000000000", 10)
	b.Div(b, big.NewInt(2))
	b.Div(b, big.NewInt(10))
	log.Print(b)
	b.Div(b, big.NewInt(4))
	log.Print(b)
	return nil
}

func (r *rewardCases) CaseCreateDelegate() error {
	ctx := context.Background()
	delAccount, _ := AccountPool.Get().(*PriAccount)
	if delAccount.Address.String() == r.params.DelegateReward.StakingAccount {
		delAccount, _ = AccountPool.Get().(*PriAccount)
	}
	nodeKey := crypto.HexMustToECDSA(r.params.DelegateReward.NodeId)
	id := discover.PubkeyID(&nodeKey.PublicKey)

	d, _ := new(big.Int).SetString("20000000000000000000", 10)
	hash, err := r.DelegateTransaction(ctx, delAccount, id, 0, d)
	if err != nil {
		return err
	}
	if err := r.WaitTransactionByHash(ctx, hash); err != nil {
		return err
	}

	log.Print("hash:", hash.String())
	log.Print("account:", delAccount.Address.String())

	receipt, err := r.client.TransactionReceipt(ctx, hash)
	if err != nil {
		return err
	}
	if len(receipt.Logs[0].Data) > 0 {
		var code [][]byte
		if err := rlp.DecodeBytes(receipt.Logs[0].Data, &code); err != nil {
			return err
		}
		log.Print("code:", string(code[0]))
	}
	r.params.DelegateReward.DelegateAccount = delAccount.Address.String()
	return nil
}

func (r *rewardCases) CaseCreateDelegateAndWithdrawDelegateReward() error {
	//ctx := context.Background()
	//delAccount, _ := AccountPool.Get().(*PriAccount)
	//if delAccount.Address.String() == r.params.DelegateReward.StakingAccount{
	//	delAccount, _ = AccountPool.Get().(*PriAccount)
	//}
	//nodeKey := crypto.HexMustToECDSA(r.params.DelegateReward.NodeId)
	//id := discover.PubkeyID(&nodeKey.PublicKey)
	//
	//d,_:=new(big.Int).SetString("20000000000000000000",10)
	//hash,err:= r.DelegateTransaction(ctx,delAccount,id,0,d)
	//if err!=nil{
	//	return err
	//}
	//if err:= r.WaitTransactionByHash(ctx,hash);err!=nil{
	//	return err
	//}
	//
	//log.Print("hash:",hash.String())
	//log.Print("account:",delAccount.Address.String())
	//
	//receipt,err:=  r.client.TransactionReceipt(ctx,hash)
	//if err!=nil{
	//	return err
	//}
	//if len(receipt.Logs[0].Data)>0{
	//	var code [][]byte
	//	if err:= rlp.DecodeBytes(	receipt.Logs[0].Data, &code);err!=nil{
	//		return err
	//	}
	//	log.Print("code:",string(code[0]))
	//}
	//
	//delAccount2, _ := AccountPool.Get().(*PriAccount)
	//if delAccount2.Address.String() == r.params.DelegateReward.StakingAccount{
	//	delAccount2, _ = AccountPool.Get().(*PriAccount)
	//}
	//
	//r.addJobs("发起第二笔委托", func(block *types.Block, params ...interface{}) (b bool, err error) {
	//
	//	hash2,err:= r.DelegateTransaction(ctx,delAccount2,id,0,d)
	//	if err!=nil{
	//		return false,err
	//	}
	//	if err:= r.WaitTransactionByHash(ctx,hash2);err!=nil{
	//		return false,err
	//	}
	//
	//	log.Print("hash2:",hash2.String())
	//	log.Print("account2:",delAccount2.Address.String())
	//
	//	receipt2,err:=  r.client.TransactionReceipt(ctx,hash2)
	//	if err!=nil{
	//		return false,err
	//	}
	//	if len(receipt2.Logs[0].Data)>0{
	//		var code [][]byte
	//		if err:= rlp.DecodeBytes(	receipt2.Logs[0].Data, &code);err!=nil{
	//			return false,err
	//		}
	//		log.Print("code2:",string(code[0]))
	//	}
	//})
	//
	//r.addJobs("领取委托收益", func(block *types.Block, params ...interface{}) (b bool, err error) {
	//
	//})

	return nil
}

func (r *rewardCases) CaseWithdrawDelegateReward() error {
	ctx := context.Background()
	pr := allAccounts[common.HexToAddress(r.params.DelegateReward.DelegateAccount)]
	hash, err := r.WithdrawDelegateReward(ctx, pr)
	if err != nil {
		return err
	}
	log.Print("hash:", hash.String())
	if err := r.WaitTransactionByHash(ctx, hash); err != nil {
		return err
	}
	receipt, err := r.client.TransactionReceipt(ctx, hash)
	if err != nil {
		return err
	}
	if len(receipt.Logs[0].Data) > 0 {
		var code [][]byte
		if err := rlp.DecodeBytes(receipt.Logs[0].Data, &code); err != nil {
			return err
		}
		log.Print("code:", string(code[0]))
		if string(code[0]) == "0" {
			var res2 []reward.NodeDelegateReward
			if err := rlp.DecodeBytes(code[1], &res2); err != nil {
				return err
			}
			log.Printf("NodeDelegateReward:%+v", res2)
		}
	}
	return nil
}

func (r *rewardCases) CaseCreateStaking() error {
	ctx := context.Background()
	stakingAccount, _ := AccountPool.Get().(*PriAccount)
	var input Dto.Staking
	err := bls.Init(int(bls.BLS12_381))
	if err != nil {
		return err
	}
	blsKey := new(bls.SecretKey)

	key, err := hex.DecodeString(r.params.DelegateReward.BlsKey)
	if err != nil {
		return err
	}
	if err = blsKey.SetLittleEndian(key); err != nil {
		return err
	}

	var keyEntries bls.PublicKeyHex
	blsHex := hex.EncodeToString(blsKey.GetPublicKey().Serialize())
	if err := keyEntries.UnmarshalText([]byte(blsHex)); err != nil {
		return err
	}

	input.BlsPubKey = keyEntries

	tmp2, err := blsKey.MakeSchnorrNIZKP()
	if err != nil {
		return err
	}
	proofByte, err := tmp2.MarshalText()
	if nil != err {
		return err
	}
	var proofHex bls.SchnorrProofHex
	if err := proofHex.UnmarshalText(proofByte); err != nil {
		return err
	}
	input.BlsProof = proofHex

	programVersion := uint32(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch)

	nodeKey := crypto.HexMustToECDSA(r.params.DelegateReward.NodeId)

	handle := node.GetCryptoHandler()
	handle.SetPrivateKey(nodeKey)
	versionSign := common.VersionSign{}
	versionSign.SetBytes(handle.MustSign(programVersion))
	input.ProgramVersion = programVersion
	input.ProgramVersionSign = versionSign

	input.Amount, _ = new(big.Int).SetString("150000000000000000000000000", 10)
	input.Typ = 0
	_, add := r.generateEmptyAccount()
	log.Print("benefitAddress:", add.String())
	input.BenefitAddress = add

	input.NodeId = discover.PubkeyID(&nodeKey.PublicKey)
	input.RewardPer = 1000
	log.Print("begin create staking")
	stakingTransaction, err := r.CreateStakingTransaction(ctx, stakingAccount, input)
	if err := r.WaitTransactionByHash(ctx, stakingTransaction); err != nil {
		return err
	}
	log.Print("tx hash:", stakingTransaction.String())
	r.params.DelegateReward.StakingAccount = stakingAccount.Address.String()

	receipt, err := r.client.TransactionReceipt(ctx, stakingTransaction)
	if err != nil {
		return err
	}
	if len(receipt.Logs[0].Data) > 0 {

		var code [][]byte
		if err := rlp.DecodeBytes(receipt.Logs[0].Data, &code); err != nil {
			return err
		}
		log.Print("code:", string(code[0]))
	}
	res := r.CallCandidateInfo(ctx, input.NodeId)
	log.Print("CandidateInfo:", res)

	tmp := res["StakingBlockNum"].(float64)

	r.params.DelegateReward.StakingNum = uint64(tmp)
	return nil
}

func (r *rewardCases) CaseQueryCandidateInfo() error {
	ctx := context.Background()
	nodeKey := crypto.HexMustToECDSA(r.params.DelegateReward.NodeId)
	id := discover.PubkeyID(&nodeKey.PublicKey)

	res := r.CallCandidateInfo(ctx, id)
	log.Print("CandidateInfo:", res)

	tmp := res["StakingBlockNum"].(float64)

	r.params.DelegateReward.StakingNum = uint64(tmp)

	return nil
}

func (r *rewardCases) CaseQueryValidatorList() error {
	ctx := context.Background()
	res3 := r.GetGetValidatorList(ctx)
	log.Print("GetValidatorList:", res3)
	log.Print("GetValidatorList,len", len(res3))
	return nil
}

func (r *rewardCases) CaseQueryVerifierList() error {
	ctx := context.Background()
	res3 := r.GetVerifierList(ctx)
	log.Print("GetVerifierList:", res3)
	log.Print("GetVerifierList,len", len(res3))
	return nil
}

func (r *rewardCases) CaseQueryDelegateInfo() error {
	ctx := context.Background()
	nodeKey := crypto.HexMustToECDSA(r.params.DelegateReward.NodeId)
	id := discover.PubkeyID(&nodeKey.PublicKey)

	res := r.GetDelegateInfo(ctx, r.params.DelegateReward.StakingNum, common.HexToAddress(r.params.DelegateReward.DelegateAccount), id)
	log.Print("GetDelegateInfo:", res)
	return nil
}

func (r *rewardCases) CaseQueryDelegateRewardInfo() error {
	ctx := context.Background()

	nodeKey := crypto.HexMustToECDSA(r.params.DelegateReward.NodeId)
	id := discover.PubkeyID(&nodeKey.PublicKey)

	res, err := r.CallGetDelegateReward(ctx, common.HexToAddress(r.params.DelegateReward.DelegateAccount), []discover.NodeID{id})
	if err != nil {
		return err
	}
	log.Print("GetDelegateReward:", res)
	return nil
}

func (r *rewardCases) CaseQueryDelegateRewardInfo333() error {
	ctx := context.Background()

	nodeKey := crypto.HexMustToECDSA(r.params.DelegateReward.NodeId)
	id := discover.PubkeyID(&nodeKey.PublicKey)

	res, err := r.CallGetDelegateReward(ctx, common.HexToAddress("0x1cb1Cf91019Af6e404206a25e488A3aA9E38f47e"), []discover.NodeID{id})
	if err != nil {
		return err
	}
	log.Print("GetDelegateReward:", res)

	res2, err := r.CallGetDelegateReward(ctx, common.HexToAddress("0x672E622CeF56fb3E9361f70bF1a47E2629375FDA"), []discover.NodeID{id})
	if err != nil {
		return err
	}


	tmp:= res.([]interface {})
	tmp3:= tmp[0].(map[string]interface{})
	reward1:= tmp3["reward"].(string)


	account1:= hexutil.MustDecodeBig(reward1)
	tmp2:= res2.([]interface {})
	tmp4:= tmp2[0].(map[string]interface{})
	reward2:= tmp4["reward"].(string)
	account2:= hexutil.MustDecodeBig(reward2)


	account3:=  r.GetBalance(ctx,vm.DelegateRewardPoolAddr,new(big.Int).SetInt64(4800))

	log.Printf("get:%v, pool:%v,left:%v",new(big.Int).Add(account1,account2),account3,new(big.Int).Sub(account3,new(big.Int).Add(account1,account2)))
	return nil
}

func (r *rewardCases) CaseQueryDelegateRewardInfo2() error {
	ctx := context.Background()
	nodes := make([]discover.NodeID, 7000)
	for i := 0; i < 6999; i++ {
		key, err := crypto.GenerateKey()
		if err != nil {
			return nil
		}
		id := discover.PubkeyID(&key.PublicKey)
		nodes[i] = id
	}
	nodeKey := crypto.HexMustToECDSA(r.params.DelegateReward.NodeId)
	id := discover.PubkeyID(&nodeKey.PublicKey)
	nodes[6999] = id

	for i := 0; i < 10; i++ {
		go func() {
			for {
				_, err := r.CallGetDelegateReward(ctx, common.HexToAddress(r.params.DelegateReward.DelegateAccount), nodes)
				if err != nil {
					log.Print(err)
				}
				time.Sleep(time.Millisecond * 100)
			}
		}()
	}

	time.Sleep(1000 * time.Second)

	return nil
}
