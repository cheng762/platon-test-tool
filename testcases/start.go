package testcases

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/PlatONnetwork/platon-test-tool/common"

	"gopkg.in/urfave/cli.v1"
)

var (
	StartCmd = cli.Command{
		Name:   "start",
		Usage:  "start all test cases",
		Action: start,
		Flags:  StartFlag,
	}
	ExecCmd = cli.Command{
		Name:   "exec",
		Usage:  "exec a test cases",
		Action: exec,
		Flags:  ExecFlag,
	}
	ListCmd = cli.Command{
		Name:   "list",
		Usage:  "list  cases",
		Action: list,
	}
)

func init() {
	allCases = make(map[string]caseTest)
	allCases["restricting"] = new(restrictCases)
	allCases["reward"] = new(rewardCases)
	allCases["staking"] = new(stakingCases)

}

type caseTest interface {
	Start() error
	Exec(string) error
	Prepare() error
	End() error
	List() []string
}

var (
	allCases  map[string]caseTest
	TxManager *common.TxManager
)

func start(c *cli.Context) error {
	var wg sync.WaitGroup
	TxManager = common.NewTxManager(c.String(ConfigPathFlag.Name))
	if err := TxManager.LoadAccounts(); err != nil {
		return err
	}

	for name, value := range allCases {
		wg.Add(1)
		go func(caseName string, caseFunc caseTest) {
			defer wg.Done()
			if err := caseFunc.Prepare(); err != nil {
				log.Printf("exec %v, Prepare fail:%v", caseName, err)
				return
			}
			if err := caseFunc.Start(); err != nil {
				log.Printf("exec %v, Start fail:%v", caseName, err)
				return
			}
			if err := caseFunc.End(); err != nil {
				log.Printf("exec %v, End fail:%v", caseName, err)
				return
			}
		}(name, value)
	}
	wg.Wait()
	if err := TxManager.SaveAccounts(); err != nil {
		return err
	}
	log.Print("all case exec done")
	return nil
}

func exec(c *cli.Context) error {

	funcName := c.String(FuncNameFlag.Name)
	caseName := c.Args().First()
	txm := common.NewTxManager(c.String(ConfigPathFlag.Name))
	if err := txm.LoadAccounts(); err != nil {
		return err
	}
	cases, ok := allCases[caseName]
	if !ok {
		return fmt.Errorf("not find the case:%v", caseName)
	}
	if err := cases.Prepare(); err != nil {
		return fmt.Errorf("exec %v, Prepare fail:%v", caseName, err)
	}
	if funcName == "" {
		if err := cases.Start(); err != nil {
			return fmt.Errorf("exec %v, Start fail:%v", caseName, err)
		}
	} else {
		if err := cases.Exec(funcName); err != nil {
			return fmt.Errorf("exec %v, Start fail:%v", funcName, err)
		}
	}
	if err := cases.End(); err != nil {
		return fmt.Errorf("exec %v, End fail", caseName)
	}
	if err := txm.SaveAccounts(); err != nil {
		return err
	}
	return nil
}

func list(c *cli.Context) error {
	caseName := c.Args().First()
	var output string
	if caseName != "" {
		cases := allCases[caseName]
		names := cases.List()
		output = strings.Join(names, ",")
		fmt.Printf("case %s support funcs: %s\n", caseName, output)
	} else {
		var names []string
		for key, _ := range allCases {
			names = append(names, key)
		}
		output = strings.Join(names, ",")
		fmt.Printf("support cases:%s\n", output)
	}
	return nil
}
