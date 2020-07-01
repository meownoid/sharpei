package main

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type ProfileConfig struct {
	Width         int    `yaml:"width"`
	Height        int    `yaml:"height"`
	InputProfile  string `yaml:"input_profile"`
	OutputProfile string `yaml:"output_profile"`
	Type          string `yaml:"type"`
	Quality       int    `yaml:"quality"`
	Compression   int    `yaml:"compression"`
}

type Config struct {
	Output   string                   `yaml:"output"`
	Format   string                   `yaml:"format"`
	Rewrite  bool                     `yaml:"rewrite"`
	Profiles map[string]ProfileConfig `yaml:"profiles"`
}

func loadConfig(filename string) (*Config, error) {
	r, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()

	content, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var cfg Config

	err = yaml.Unmarshal(content, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
