package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"sync"
)

// Config struct to hold the configuration data
type Config struct {
	GroupingElements []string `yaml:"grouping_elements"`
}

var (
	// globalConfig is the instance of Config that will be used globally
	globalConfig *Config
	// mutex to ensure thread-safe access to globalConfig
	mutex sync.Mutex
)

// LoadConfig loads the configuration from a YAML file into the globalConfig variable
func LoadConfig(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading config file: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("error unmarshaling config: %v", err)
	}

	// Safely assign the config to the globalConfig variable
	mutex.Lock()
	globalConfig = &config
	mutex.Unlock()

	return nil
}

// GetConfig returns the globalConfig instance
func GetConfig() *Config {
	mutex.Lock()
	defer mutex.Unlock()
	return globalConfig
}
