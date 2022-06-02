package config

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Params map[string]interface{} `yaml:"Params"`
	Menu   map[string]interface{} `yaml:"Menu"`
	Theme  string                 `yaml:"Theme"`
	Port   int                    `yaml:"Port"`
	Forge  struct {
		Path  string
		Debug bool
	}
}

func Get() (*Config, error) {
	var config Config
	b, err := ioutil.ReadFile("config.yml")
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(b, &config); err != nil {
		return nil, err
	}
	if os.Getenv("DEBUG") != "" {
		config.Forge.Debug = true
	}
	return &config, nil
}
