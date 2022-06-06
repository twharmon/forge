package config

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Theme struct {
		Name   string                 `yaml:"Name"`
		Params map[string]interface{} `yaml:"Params"`
	} `yaml:"Theme"`

	Markdown struct {
		Extensions []string `yaml:"Extensions"`
	} `yaml:"Markdown"`

	DevServer struct {
		Port int `yaml:"Port"`
	} `yaml:"DevServer"`

	Forge struct {
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
