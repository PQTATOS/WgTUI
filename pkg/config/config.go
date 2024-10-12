package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type AppConfig struct {
	VPN VPNConfig `yaml:"vpn"`
	Bot BotConfig `yaml:"bot"`
}

type VPNConfig struct {
	Repo struct {
		ConfigSaveDir string `yaml:"config_save_dir"`
		ServerConfigName string `yaml:"server_config_name"`
	} `yaml:"repo"`
}

type BotConfig struct {
	Token string `yaml:"token"`
}

var Cfg AppConfig

func ReadConfig() error {
	f, err := os.Open("config.yml")
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&Cfg); err != nil {
		return err
	}
	return nil
}