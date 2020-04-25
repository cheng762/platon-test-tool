package common

import "github.com/PlatONnetwork/PlatON-Go/common"

const (
	DefaultConfigFilePath = "/config.json"
)

type Config struct {
	Account                common.Address    `json:"account"` //address in genesis
	GeinisPrikey           string            `json:"prikey"`
	Url                    string            `json:"url"`
	Dir                    string            `json:"dir"`
	TestCaseConfig         map[string]string `json:"test_cases_config"`
	PrivateKeyFile         string            `json:"private_key_file"`
	DefaultAccountAddrFile string            `json:"default_account_addr_file"`
}
