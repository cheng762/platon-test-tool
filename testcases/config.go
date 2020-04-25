package testcases

import (
	"encoding/json"
	"fmt"
	"github.com/PlatONnetwork/PlatON-Go/common"
	"io/ioutil"
	"os"
	"path/filepath"
)

var (
	config = Config{}
)

func parseConfigJson(configPath string) {
	if configPath == "" {
		dir, _ := os.Getwd()
		configPath = dir + DefaultConfigFilePath
	}

	if !filepath.IsAbs(configPath) {
		configPath, _ = filepath.Abs(configPath)
	}

	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(fmt.Errorf("parse config file error,%s", err.Error()))
	}

	if err := json.Unmarshal(bytes, &config); err != nil {
		panic(fmt.Errorf("parse config to json error,%s", err.Error()))
	}
}

func loadData(configPath string) {
	parseConfigJson(configPath)
	parsePkFile()
	for _, value := range allAccounts {
		AccountPool.Put(value)
	}
}
