package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"qc-process/internal/models"
)

type ItemConfig struct {
	Name         string  `yaml:"name"`
	StandardValue float64 `yaml:"standard_value"`
	ToleranceMin float64 `yaml:"tolerance_min"`
	ToleranceMax float64 `yaml:"tolerance_max"`
}

type YAMLConfig struct {
	Items []ItemConfig `yaml:"items"`
}

func LoadItemsFromYAML(filePath string) ([]models.InspectionItem, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("错误：无法读取配置文件: %v", err)
	}

	var config YAMLConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("错误：YAML 格式错误: %v", err)
	}

	var items []models.InspectionItem
	for _, item := range config.Items {
		items = append(items, models.InspectionItem{
			Name:          item.Name,
			StandardValue:  item.StandardValue,
			ToleranceMin:  item.ToleranceMin,
			ToleranceMax:  item.ToleranceMax,
		})
	}

	return items, nil
}
