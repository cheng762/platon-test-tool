package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func LoadConfig(path string, config interface{}) error {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("parse config file error,%s", err.Error())
	}

	if err := json.Unmarshal(bytes, config); err != nil {
		return fmt.Errorf("parse config to json error,%s", err.Error())
	}
	return nil
}

func SaveConfig(path string, config interface{}) error {
	byts, err := json.MarshalIndent(config, "", "\t")
	_, err = os.Create(path)
	if err != nil {
		return fmt.Errorf("create addr.json error%s \n", err.Error())
	}
	err = ioutil.WriteFile(path, byts, 0644)
	if err != nil {
		return fmt.Errorf("write to addr.json error%s \n", err.Error())
	}
}
