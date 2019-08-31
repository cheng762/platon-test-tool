package testcases

import (
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"log"
	"strings"
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
		Usage:  "list all cases",
		Action: list,
		Flags:  ListFlag,
	}
)

func init() {
	allCases = make(map[string]caseTest)
	allCases["restricting"] = new(restrictCases)
	allCases["init_token"] = new(initCases)
}

type caseTest interface {
	Start() error
	Exec(string) error
	Prepare() error
	End() error
	List() []string
}

var allCases map[string]caseTest

func start(c *cli.Context) {
	parseConfigJson(c.String(ConfigPathFlag.Name))
	for name, value := range allCases {
		if err := value.Prepare(); err != nil {
			log.Printf("exec %v, Prepare fail:%v", name, err)
			return
		}
		if err := value.Start(); err != nil {
			log.Printf("exec %v, Start fail:%v", name, err)
			return
		}
		if err := value.End(); err != nil {
			log.Printf("exec %v, End fail:%v", name, err)
			return
		}
	}
}

func exec(c *cli.Context) {
	parseConfigJson(c.String(ConfigPathFlag.Name))
	funcName := c.String(FuncNameFlag.Name)
	caseName := c.String(CaseNameFlag.Name)
	cases := allCases[caseName]
	if err := cases.Prepare(); err != nil {
		log.Printf("exec %v, Prepare fail:%v", funcName, err)
		return
	}
	if funcName == "" {
		if err := cases.Start(); err != nil {
			log.Printf("exec %v, Start fail:%v", funcName, err)
			return
		}
	} else {
		if err := cases.Exec(funcName); err != nil {
			log.Printf("exec %v, Start fail:%v", funcName, err)
			return
		}
	}
	if err := cases.End(); err != nil {
		log.Printf("exec %v, End fail", funcName)
		return
	}
}

func list(c *cli.Context) {
	caseName := c.String(CaseNameFlag.Name)
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
}
