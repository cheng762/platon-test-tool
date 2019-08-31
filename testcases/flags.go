package testcases

import "gopkg.in/urfave/cli.v1"

var (
	AccountSizeFlag = cli.IntFlag{
		Name:  "size",
		Value: 10,
		Usage: "account size",
	}
	TransferValueFlag = cli.StringFlag{
		Name:  "value",
		Value: "0x200000000000000000000000000000000000000000000000000000000", //one
		Usage: "transfer value",
	}
	ConfigPathFlag = cli.StringFlag{
		Name:  "config,c",
		Usage: "config path",
	}
	FuncNameFlag = cli.StringFlag{
		Name:  "func,f",
		Usage: "use specic func ",
	}
	StartFlag = []cli.Flag{
		ConfigPathFlag,
	}
	ExecFlag = []cli.Flag{
		FuncNameFlag,
		ConfigPathFlag,
	}
	stabPrepareCmdFlags = []cli.Flag{
		AccountSizeFlag,
		TransferValueFlag,
		ConfigPathFlag,
	}
)
