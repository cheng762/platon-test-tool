package testcases

import (
	"fmt"
	"github.com/PlatONnetwork/PlatON-Go/crypto"
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

func prepareAccount(c *cli.Context) {
	size := c.Int(AccountSizeFlag.Name)
	value := c.String(TransferValueFlag.Name)

	parseConfigJson(c.String(ConfigPathFlag.Name))

	err := PrepareAccount(size, value)
	if err != nil {
		panic(fmt.Errorf("send raw transaction error,%s", err.Error()))
	}
}
