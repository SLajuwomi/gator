package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func (c Config) SetUser(userName string) error {
	c.CurrentUserName = userName
	marshaledConfig, err := json.Marshal(c)
	if err != nil {
		return err
	}
	configFilePath, err := GetConfigFilePath()
	if err != nil {
		return err
	}
	err = os.WriteFile(configFilePath, marshaledConfig, 0666)
	if err != nil {
		return err
	}
	return nil
}
