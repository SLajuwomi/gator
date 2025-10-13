package config

import (
	"encoding/json"
	"log"
	"os"
)

const configFileName = "/.gatorconfig.json"

func GetConfigFilePath() (string, error) {
	homeDirPath, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configFilePath := homeDirPath + configFileName
	return configFilePath, nil
}

func Read() Config {
	var configStruct Config
	configFilePath, err := GetConfigFilePath()
	if err != nil {
		log.Fatal(err)
	}
	configContent, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Fatal(err)
	}
	if err = json.Unmarshal(configContent, &configStruct); err != nil {
		log.Fatal(err)
	}
	return configStruct
}
