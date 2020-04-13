package Factory

import (
	"crypto/ecdsa"
	"github.com/PlatONnetwork/PlatON-Go/common"
	"github.com/PlatONnetwork/PlatON-Go/crypto"
)

//生成一个新的空账户，私钥和地址
func GenerateEmptyAccount() (*ecdsa.PrivateKey, common.Address) {
	privateKey, _ := crypto.GenerateKey()
	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	return privateKey, address
}
