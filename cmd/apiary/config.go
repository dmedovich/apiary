package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// config mirrors an optional apiary.yaml file in the working directory, so
// projects can pin their settings instead of passing long flag lists (handy
// with `//go:generate apiary`). Explicitly-set CLI flags override these values.
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

// configFiles are the names looked up (in order) in the current directory.
var configFiles = []string{"apiary.yaml", ".apiary.yaml"}

// loadConfig reads the first config file that exists, or returns a zero config
// when none is present. A malformed config is a fatal error.
func loadConfig() config {
	for _, name := range configFiles {
		data, err := os.ReadFile(name)
		if err != nil {
			continue // not present / unreadable — try the next
		}
		var cfg config
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			log.Fatalf("apiary: parse %s: %v", name, err)
		}
		return cfg
	}
	return config{}
}
