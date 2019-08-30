package testcases

import (
	"crypto/ecdsa"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/PlatONnetwork/PlatON-Go/common"
	"github.com/PlatONnetwork/PlatON-Go/crypto"
	"github.com/PlatONnetwork/PlatON-Go/crypto/secp256k1"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
)

var (
	AccountPool = make(map[common.Address]*PriAccount)
)

type PriAccount struct {
	Priv  *ecdsa.PrivateKey
	Nonce uint64
}

func GetAllAddress() []common.Address {
	addrsPath := path.Join(config.Dir, config.DefaultAccountAddrFile)
	bytes, err := ioutil.ReadFile(addrsPath)
	if err != nil {
		panic(fmt.Errorf("get all address array error,%s \n", err.Error()))
	}
	var addrs []common.Address
	err = json.Unmarshal(bytes, &addrs)
	if err != nil {
		panic(fmt.Errorf("parse address to array error,%s \n", err.Error()))
	}

	return addrs
}

func GetRandomAddr(addrs []common.Address) (common.Address, common.Address) {
	if len(addrs) == 0 {
		return common.ZeroAddr, common.ZeroAddr
	}
	fromIndex := rand.Intn(len(addrs))
	toIndex := rand.Intn(len(addrs))
	for toIndex == fromIndex {
		toIndex = rand.Intn(len(addrs))
	}
	return addrs[fromIndex], addrs[toIndex]
}

func generateAccount(size int) {
	addrs := make([]common.Address, size)
	for i := 0; i < size; i++ {
		privateKey, _ := crypto.GenerateKey()
		address := crypto.PubkeyToAddress(privateKey.PublicKey)
		AccountPool[address] = &PriAccount{privateKey, 0}
		addrs[i] = address
	}
	savePrivateKeyPool()
	saveAddrs(addrs)
}

func savePrivateKeyPool() {
	pkFile := path.Join(config.Dir, config.PrivateKeyFile)
	gob.Register(&secp256k1.BitCurve{})
	file, err := os.Create(pkFile)
	if err != nil {
		panic(fmt.Errorf("save private key err,%s,%s", pkFile, err.Error()))
	}
	os.Truncate(pkFile, 0)
	enc := gob.NewEncoder(file)
	err = enc.Encode(AccountPool)
	if err != nil {
		panic(err.Error())
	}
}

func saveAddrs(addrs []common.Address) {
	addrsPath := path.Join(config.Dir, config.DefaultAccountAddrFile)
	os.Truncate(addrsPath, 0)
	byts, err := json.MarshalIndent(addrs, "", "\t")
	_, err = os.Create(addrsPath)
	if err != nil {
		panic(fmt.Errorf("create addr.json error%s \n", err.Error()))
	}
	err = ioutil.WriteFile(addrsPath, byts, 0644)
	if err != nil {
		panic(fmt.Errorf("write to addr.json error%s \n", err.Error()))
	}
}

func parsePkFile() {
	pkFile := path.Join(config.Dir, config.PrivateKeyFile)
	gob.Register(&secp256k1.BitCurve{})
	file, err := os.Open(pkFile)
	if err != nil {
		panic(err)
	}
	dec := gob.NewDecoder(file)
	err2 := dec.Decode(&AccountPool)
	if err2 != nil {
		panic(err2)
	}
}
