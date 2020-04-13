package common

import (
	"github.com/PlatONnetwork/PlatON-Go/common"
	"github.com/PlatONnetwork/PlatON-Go/common/vm"
	"github.com/PlatONnetwork/PlatON-Go/rlp"
)

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

func funcTypeToContractAddress(funcType uint16) common.Address {
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
	return toadd
}
