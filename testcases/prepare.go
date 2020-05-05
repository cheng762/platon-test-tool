package testcases

import (
	"github.com/PlatONnetwork/platon-test-tool/common"
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

func prepareAccount(c *cli.Context) error {
	size := c.Int(AccountSizeFlag.Name)
	value := c.String(TransferValueFlag.Name)
	TxManager = common.NewTxManager(c.String(ConfigPathFlag.Name))
	if err := TxManager.LoadAccounts(); err != nil {
		return err
	}
	if err := TxManager.PrepareAccount(size, value); err != nil {
		return err
	}
	if err := TxManager.SaveAccounts(); err != nil {
		return err
	}
	return nil
}
