package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type config struct {
	Title       string   `yaml:"title"`
	Version     string   `yaml:"version"`
	Description string   `yaml:"description"`
	Out         string   `yaml:"out"`
	Format      string   `yaml:"format"`
	Security    []string `yaml:"security"`
	Servers     []string `yaml:"servers"`
	Patterns    []string `yaml:"patterns"`
}

var configFiles = []string{"apiary.yaml", ".apiary.yaml"}

func loadConfig() config {
	for _, name := range configFiles {
		data, err := os.ReadFile(name)
		if err != nil {
			continue
		}
		var cfg config
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			log.Fatalf("apiary: parse %s: %v", name, err)
		}
		return cfg
	}
	return config{}
}
