package config

type StakingCaseConfig struct {
	MergeDelegate *StakingConfig
}

type StakingConfig struct {
	BlsKey    string `json:"bls_key"`
	NodeKey   string `json:"node_key"`
	RewardPer uint16 `json:"reward_per"`
	Typ       uint16 `json:"typ"`
	NodeUrl   string `json:"node_url"`
}
