package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	RemoteAreas []string            `json:"remote_areas"`
	AddressDB   map[string]Address  `json:"address_db"`
	ZoneRules   ZoneRules           `json:"zone_rules"`
}

type Address struct {
	Province string `json:"province"`
	City     string `json:"city"`
	District string `json:"district"`
}

type ZoneRules struct {
	Local   int `json:"local"`
	Intra   int `json:"intra"`
	Neighbor int `json:"neighbor"`
	Inter   int `json:"inter"`
	Remote  int `json:"remote"`
}

var instance *Config

func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("配置文件读取失败: %v", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("配置文件解析失败: %v", err)
	}

	instance = &cfg
	return nil
}

func Get() *Config {
	return instance
}
