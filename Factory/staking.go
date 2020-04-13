package Factory

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/PlatONnetwork/PlatON-Go/common"
	"github.com/PlatONnetwork/PlatON-Go/crypto"
	"github.com/PlatONnetwork/PlatON-Go/crypto/bls"
	"github.com/PlatONnetwork/PlatON-Go/node"
	"github.com/PlatONnetwork/PlatON-Go/p2p/discover"
	"github.com/PlatONnetwork/PlatON-Go/params"
	"github.com/PlatONnetwork/platon-test-tool/Dto"
	"io/ioutil"
	"math/big"
)

type StakingConfig struct {
	BlsKey    string `json:"bls_key"`
	NodeKey   string `json:"node_key"`
	RewardPer uint16 `json:"reward_per"`
	Typ       uint16 `json:"typ"`
}

func LoadStakingConfig(path string) *StakingConfig {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("parse staking config file error,%s", err.Error()))
	}
	config := new(StakingConfig)
	if err := json.Unmarshal(bytes, &config); err != nil {
		panic(fmt.Errorf("parse restrictCases config to json error,%s", err.Error()))
	}
	return config
}

func BuildStaking(config *StakingConfig, stakingAmount *big.Int) (*Dto.Staking, error) {
	input := new(Dto.Staking)
	err := bls.Init(int(bls.BLS12_381))
	if err != nil {
		return nil, err
	}
	blsKey := new(bls.SecretKey)

	key, err := hex.DecodeString(config.BlsKey)
	if err != nil {
		return nil, err
	}
	if err = blsKey.SetLittleEndian(key); err != nil {
		return nil, err
	}

	var keyEntries bls.PublicKeyHex
	blsHex := hex.EncodeToString(blsKey.GetPublicKey().Serialize())
	if err := keyEntries.UnmarshalText([]byte(blsHex)); err != nil {
		return nil, err
	}

	input.BlsPubKey = keyEntries

	tmp2, err := blsKey.MakeSchnorrNIZKP()
	if err != nil {
		return nil, err
	}
	proofByte, err := tmp2.MarshalText()
	if nil != err {
		return nil, err
	}
	var proofHex bls.SchnorrProofHex
	if err := proofHex.UnmarshalText(proofByte); err != nil {
		return nil, err
	}
	input.BlsProof = proofHex

	programVersion := uint32(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch)

	nodeKey := crypto.HexMustToECDSA(config.NodeKey)

	handle := node.GetCryptoHandler()
	handle.SetPrivateKey(nodeKey)
	versionSign := common.VersionSign{}
	versionSign.SetBytes(handle.MustSign(programVersion))
	input.ProgramVersion = programVersion
	input.ProgramVersionSign = versionSign

	input.Amount = new(big.Int).Set(stakingAmount)

	input.Typ = config.Typ
	_, add := GenerateEmptyAccount()
	input.BenefitAddress = add
	input.NodeId = discover.PubkeyID(&nodeKey.PublicKey)
	input.RewardPer = config.RewardPer
	return input, nil
}
