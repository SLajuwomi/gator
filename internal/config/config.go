package config

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func (c Config) SetUser(userName string) {
	c.CurrentUserName = userName
	marshaledConfig, err := json.Marshal(c)
	if err != nil {
		log.Fatal(err)
	}
	configFilePath, err := GetConfigFilePath()
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(configFilePath, marshaledConfig, 0666)
	if err != nil {
		log.Fatal(err)
	}
}
