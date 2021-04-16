package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type FlowConfig struct {
	Accounts  map[string]FlowConfigAccount `json:"accounts"`
	Contracts map[string]string            `json:"contracts"`
}

type FlowConfigAccount struct {
	Address  string `json:"address"`
	Keys     string `json:"keys"`
	SigAlgo  string `json:"sigAlgorithm"`
	HashAlgo string `json:"hashAlgorithm"`
}

func ReadFile(configPath string) FlowConfig {
	jsonFile, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Config file not found at %s \n", configPath)
		} else {
			fmt.Printf("Failed to load config from %s: %s\n", configPath, err.Error())
		}
		os.Exit(1)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var conf FlowConfig
	err = json.Unmarshal(byteValue, &conf)
	Handle(err)

	// Add default values
	for key, value := range conf.Accounts {
		if value.SigAlgo == "" {
			value.SigAlgo = "ECDSA_P256"
		}
		if value.HashAlgo == "" {
			value.HashAlgo = "SHA3_256"
		}
		conf.Accounts[key] = value
	}

	return conf
}
