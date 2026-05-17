package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	SignalServer string `yaml:"signal_server"`
	MpvSocket    string `yaml:"mpv_socket"`
	DisplayName  string `yaml:"display_name"`
	LocalMode    bool   `yaml:"local_mode"`
}

func DefaultConfig() Config {
	return Config{
		SignalServer: "http://localhost:8080",
		MpvSocket:    "/tmp/tsuna-mpv.sock",
		DisplayName:  "",
		LocalMode:    false,
	}
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".tsuna.yaml"), nil
}

func Load() Config {
	cfg := DefaultConfig()

	path, err := configPath()
	if err != nil {
		return cfg
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}

	_ = yaml.Unmarshal(data, &cfg)
	return cfg
}

func Save(cfg Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
