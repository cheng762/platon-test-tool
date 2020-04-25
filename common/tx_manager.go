package common

import (
	"fmt"
	"github.com/PlatONnetwork/PlatON-Go/common"
	"github.com/PlatONnetwork/PlatON-Go/crypto"
	"sync"
)

type TxManager struct {
	nodes       map[string]*node
	allAccounts map[common.Address]*PriAccount
	accountPool sync.Pool
}

func (txm *TxManager) AddNode(url string) *node {
	node := NewNode(url)
	txm.nodes[url] = node
	return node
}

func (txm *TxManager) GetNode(url string) *node {
	node, ok := txm.nodes[url]
	if !ok {
		return txm.AddNode(url)
	}
	return node
}

func (txm *TxManager) PrepareAccount(size int, value string) error {
	generateAccount(size)
	pri, err := crypto.HexToECDSA(config.GeinisPrikey)
	if err != nil {
		return fmt.Errorf("hex to ecdsa fail:%v", err)
	}
	configAccount := &common.PriAccount{pri, 0, config.Account, nil}
	txm := new(common.TxManager)
	node := txm.AddNode(config.Url)
	for addr := range allAccounts {
		hash, err := node.SendTraction(configAccount, addr, value, nil)
		if err != nil {
			return fmt.Errorf("prepare error,send from coinbase error,%s", err.Error())
		}
		fmt.Printf("transfer hash: %s \n", hash.String())
	}
	fmt.Printf("prepare %d account finish...", size)
	return nil
}

func generateAccount(size int) {
	addrs := make([]common.Address, size)
	for i := 0; i < size; i++ {
		privateKey, _ := crypto.GenerateKey()
		address := crypto.PubkeyToAddress(privateKey.PublicKey)
		allAccounts[address] = &PriAccount{privateKey, 0, address, nil}
		addrs[i] = address
	}
	savePrivateKeyPool()
	saveAddrs(addrs)
}
