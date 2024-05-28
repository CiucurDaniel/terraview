package config

import (
	"fmt"
	"io/ioutil"
	"sync"

	"gopkg.in/yaml.v3"
)

// Resource defines a resource with its attributes
type Resource struct {
	Name       string   `yaml:"resource"`
	Attributes []string `yaml:"attributes"`
}

// Config struct to hold the configuration data
type Config struct {
	GroupingElements    []string   `yaml:"grouping_elements"`
	ImportantAttributes []Resource `yaml:"important_attributes"`
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

// PrintImportantAttributes prints the important attributes from the loaded configuration
func PrintImportantAttributes() {
	config := GetConfig()
	if config == nil {
		fmt.Println("Config not loaded")
		return
	}

	fmt.Println("Important Attributes:")
	for _, resource := range config.ImportantAttributes {
		fmt.Printf("Resource: %s\n", resource.Name)
		fmt.Println("Attributes:")
		for _, attribute := range resource.Attributes {
			fmt.Printf("- %s\n", attribute)
		}
	}
}
