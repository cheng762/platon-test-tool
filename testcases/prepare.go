package testcases

import (
	"context"
	"fmt"
	"github.com/PlatONnetwork/PlatON-Go/crypto"
	"github.com/PlatONnetwork/PlatON-Go/ethclient"
	"gopkg.in/urfave/cli.v1"
)

var (
	StabPrepareCmd = cli.Command{
		Name:   "prepare",
		Usage:  "prepare some accounts are used for  test ",
		Action: prepareAccount,
		Flags:  stabPrepareCmdFlags,
	}
)

func prepareAccount(c *cli.Context) {
	size := c.Int(AccountSizeFlag.Name)
	value := c.String(TransferValueFlag.Name)

	parseConfigJson(c.String(ConfigPathFlag.Name))

	err := PrepareAccount(size, value)
	if err != nil {
		panic(fmt.Errorf("send raw transaction error,%s", err.Error()))
	}
}

func PrepareAccount(size int, value string) error {
	if len(AccountPool) == 0 {
		generateAccount(size)
	}
	pri, err := crypto.HexToECDSA(config.GeinisPrikey)
	if err != nil {
		return fmt.Errorf("hex to ecdsa fail:%v", err)
	}
	AccountPool[config.Account] = &PriAccount{pri, 0}
	client, err := ethclient.Dial(config.Url)
	if err != nil {
		return err
	}
	for addr := range AccountPool {
		hash, err := SendRawTransaction(context.Background(), client, config.Account, addr, value, nil, "")
		if err != nil {
			return fmt.Errorf("prepare error,send from coinbase error,%s", err.Error())
		}
		fmt.Printf("transfer hash: %s \n", hash.String())
	}
	fmt.Printf("prepare %d account finish...", size)
	return nil
}
