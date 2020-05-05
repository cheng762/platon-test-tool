package common

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/PlatONnetwork/PlatON-Go/common"
	"github.com/PlatONnetwork/PlatON-Go/crypto"
)

func NewTxManager(configPath string) *TxManager {
	txm := new(TxManager)
	txm.loadConfig(configPath)
	return txm
}

type TxManager struct {
	nodes       map[string]*Node
	allAccounts map[common.Address]*PriAccount
	accountPool sync.Pool
	Config      *Config
}

func (txm *TxManager) AddNode(url string) *Node {
	node := NewNode(url, txm.Config.ChainID)
	txm.nodes[url] = node
	return node
}

func (txm *TxManager) GetNode(url string) *Node {
	node, ok := txm.nodes[url]
	if !ok {
		return txm.AddNode(url)
	}
	return node
}

func (txm *TxManager) GetRandomNode() *Node {
	for _, i2 := range txm.nodes {
		return i2
	}
	return nil
}

func (txm *TxManager) PrepareAccount(size int, value string) error {
	if value == "0" {
		return errors.New("prepareAccount fail:Account value must > 0 ")
	}
	addrs := txm.GenerateAccount(size)

	pri, err := crypto.HexToECDSA(txm.Config.GeinisPrikey)
	if err != nil {
		return fmt.Errorf("hex to ecdsa fail:%v", err)
	}
	configAccount := &PriAccount{pri, 0, crypto.PubkeyToAddress(pri.PublicKey), nil}
	node := txm.GetNode(txm.Config.RpcUrl[0])
	for _, addr := range addrs {

		hash, err := node.SendTraction(configAccount, addr.Address, value, nil)
		if err != nil {
			return fmt.Errorf("prepare error,send from coinbase error,%s", err.Error())
		}
		fmt.Printf("transfer hash: %s \n", hash.String())

		txm.allAccounts[addr.Address] = addr
	}
	txm.allAccounts[configAccount.Address] = configAccount

	fmt.Printf("prepare %d account finish...", size)
	return nil
}

func (txm *TxManager) GenerateAccount(size int) []*PriAccount {
	addrs := make([]*PriAccount, size)
	for i := 0; i < size; i++ {
		privateKey, _ := crypto.GenerateKey()
		address := crypto.PubkeyToAddress(privateKey.PublicKey)
		addrs[i] = &PriAccount{privateKey, 0, address, nil}
	}
	return addrs
}

func (txm *TxManager) GetAccount() *PriAccount {
	a := txm.accountPool.Get()
	if a == nil {
		return nil
	} else {
		return a.(*PriAccount)
	}
}

func (txm *TxManager) SaveAccounts() error {
	pkFile := txm.Config.AccountFilePath()
	return SaveConfig(pkFile, txm.allAccounts)
}

func (txm *TxManager) LoadAccounts() error {
	pkFile := txm.Config.AccountFilePath()
	if err := LoadConfig(pkFile, &txm.allAccounts); err != nil {
		return err
	}
	for _, account := range txm.allAccounts {
		txm.accountPool.Put(account)
	}
	return nil
}

func (txm *TxManager) loadConfig(configPath string) {
	if configPath == "" {
		dir, _ := os.Getwd()
		configPath = dir + DefaultConfigFilePath
	}

	if !filepath.IsAbs(configPath) {
		configPath, _ = filepath.Abs(configPath)
	}
	if err := LoadConfig(configPath, &txm.Config); err != nil {
		panic(err)
	}
}
