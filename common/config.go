package common

import "path"

const (
	DefaultConfigFilePath = "/config.json"
	DefaultAccountFile    = "account.json"
)

type Config struct {
	GeinisPrikey       string            `json:"prikey"`
	RpcUrl             []string          `json:"rpc_url"`
	ConfigDir          string            `json:"config_dir"`
	TestCaseConfigPath map[string]string `json:"test_cases_config"`
	ChainID            uint64            `json:"chain_id"`
}

func (c *Config) AccountFilePath() string {
	return path.Join(c.ConfigDir, DefaultAccountFile)
}

func (c *Config) GetCaseConfigPath(s string) string {
	return c.TestCaseConfigPath[s]
}
