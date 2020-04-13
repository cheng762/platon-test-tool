package Dto

import (
	"github.com/PlatONnetwork/PlatON-Go/common"
	"github.com/PlatONnetwork/PlatON-Go/crypto/bls"
	"github.com/PlatONnetwork/PlatON-Go/p2p/discover"
	"math/big"
)

type Staking struct {
	Typ            uint16
	BenefitAddress common.Address
	NodeId         discover.NodeID
	ExternalId     string
	NodeName       string
	Website        string
	Details        string
	Amount         *big.Int
	RewardPer uint16
	ProgramVersion     uint32
	ProgramVersionSign common.VersionSign
	BlsPubKey      bls.PublicKeyHex
	BlsProof      bls.SchnorrProofHex
}