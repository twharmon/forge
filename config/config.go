package config

import (
	"fmt"
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

func (c *Config) DevServerPort() string {
	if c.DevServer.Port == 0 {
		return ":8000"
	}
	return fmt.Sprintf(":%d", c.DevServer.Port)
}

func (c *Config) DevServerUrl() string {
	return fmt.Sprintf("http://localhost%s", c.DevServerPort())
}
