package config

import (
	"fmt"
	"os"
	"time"

	"monitor-alert/internal/types"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Global struct {
		CollectInterval time.Duration `yaml:"collect_interval"`
		EvaluateInterval time.Duration `yaml:"evaluate_interval"`
	} `yaml:"global"`
	Storage struct {
		DataDir   string `yaml:"data_dir"`
		Retention string `yaml:"retention"`
	} `yaml:"storage"`
	Rules    []types.AlertRule            `yaml:"rules"`
	Channels []map[string]interface{}     `yaml:"channels"`
}

type RawChannel struct {
	Name   string                 `yaml:"name"`
	Type   string                 `yaml:"type"`
	Config map[string]interface{} `yaml:"config"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("配置文件不存在: %s", path)
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}
	
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("配置文件格式错误: %w", err)
	}
	
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}
	
	return &cfg, nil
}

func validateConfig(cfg *Config) error {
	if cfg.Global.CollectInterval <= 0 {
		cfg.Global.CollectInterval = 10 * time.Second
	}
	if cfg.Global.EvaluateInterval <= 0 {
		cfg.Global.EvaluateInterval = 30 * time.Second
	}
	if cfg.Storage.DataDir == "" {
		cfg.Storage.DataDir = "./data"
	}
	
	if len(cfg.Rules) == 0 {
		return fmt.Errorf("未配置任何告警规则")
	}
	
	return nil
}

func ParseChannels(rawChannels []map[string]interface{}) ([]types.NotificationChannel, error) {
	channels := make([]types.NotificationChannel, 0, len(rawChannels))
	names := make(map[string]bool)
	
	for i, raw := range rawChannels {
		name, ok := raw["name"].(string)
		if !ok {
			return nil, fmt.Errorf("渠道 %d 缺少 name 字段", i)
		}
		if name == "" {
			return nil, fmt.Errorf("渠道 %d 的 name 不能为空", i)
		}
		if names[name] {
			return nil, fmt.Errorf("渠道名称重复: %s", name)
		}
		names[name] = true
		
		chType, ok := raw["type"].(string)
		if !ok {
			return nil, fmt.Errorf("渠道 '%s' 缺少 type 字段", name)
		}
		
		var typedChan types.NotificationChannel
		typedChan.Name = name
		
		switch types.ChannelType(chType) {
		case types.ChannelStdout:
			typedChan.Type = types.ChannelStdout
			typedChan.Config = struct{}{}
		case types.ChannelEmail:
			typedChan.Type = types.ChannelEmail
			cfg, err := parseEmailConfig(name, raw)
			if err != nil {
				return nil, err
			}
			typedChan.Config = cfg
		case types.ChannelWebhook:
			typedChan.Type = types.ChannelWebhook
			cfg, err := parseWebhookConfig(name, raw)
			if err != nil {
				return nil, err
			}
			typedChan.Config = cfg
		default:
			return nil, fmt.Errorf("渠道 '%s' 类型无效: %s (有效值: stdout, email, webhook)", name, chType)
		}
		
		channels = append(channels, typedChan)
	}
	
	return channels, nil
}

func parseEmailConfig(name string, raw map[string]interface{}) (types.EmailConfig, error) {
	configRaw, ok := raw["config"].(map[string]interface{})
	if !ok {
		return types.EmailConfig{}, fmt.Errorf("渠道 '%s' 缺少 config 字段", name)
	}
	
	cfg := types.EmailConfig{}
	
	if cfg.SMTPHost, ok = configRaw["smtp_host"].(string); !ok {
		return cfg, fmt.Errorf("渠道 '%s' 的 email 配置缺少 smtp_host", name)
	}
	if smtpPort, ok := configRaw["smtp_port"].(int); ok {
		cfg.SMTPPort = smtpPort
	} else if smtpPort, ok := configRaw["smtp_port"].(float64); ok {
		cfg.SMTPPort = int(smtpPort)
	} else {
		return cfg, fmt.Errorf("渠道 '%s' 的 email 配置缺少 smtp_port", name)
	}
	if cfg.From, ok = configRaw["from"].(string); !ok {
		return cfg, fmt.Errorf("渠道 '%s' 的 email 配置缺少 from", name)
	}
	
	toRaw, ok := configRaw["to"].([]interface{})
	if !ok {
		return cfg, fmt.Errorf("渠道 '%s' 的 email 配置缺少 to 数组", name)
	}
	for _, addr := range toRaw {
		if s, ok := addr.(string); ok {
			cfg.To = append(cfg.To, s)
		}
	}
	if len(cfg.To) == 0 {
		return cfg, fmt.Errorf("渠道 '%s' 的 email 配置 to 不能为空", name)
	}
	
	if username, ok := configRaw["username"].(string); ok {
		cfg.Username = username
	}
	if password, ok := configRaw["password"].(string); ok {
		cfg.Password = password
	}
	
	return cfg, nil
}

func parseWebhookConfig(name string, raw map[string]interface{}) (types.WebhookConfig, error) {
	configRaw, ok := raw["config"].(map[string]interface{})
	if !ok {
		return types.WebhookConfig{}, fmt.Errorf("渠道 '%s' 缺少 config 字段", name)
	}
	
	cfg := types.WebhookConfig{
		Timeout: 10 * time.Second,
		Method:  "POST",
	}
	
	if cfg.URL, ok = configRaw["url"].(string); !ok {
		return cfg, fmt.Errorf("渠道 '%s' 的 webhook 配置缺少 url", name)
	}
	if cfg.URL == "" {
		return cfg, fmt.Errorf("渠道 '%s' 的 webhook 配置 url 不能为空", name)
	}
	
	if method, ok := configRaw["method"].(string); ok && method != "" {
		cfg.Method = method
	}
	
	if timeoutRaw, ok := configRaw["timeout"].(string); ok {
		if d, err := time.ParseDuration(timeoutRaw); err == nil {
			cfg.Timeout = d
		}
	}
	
	if headersRaw, ok := configRaw["headers"].(map[string]interface{}); ok {
		cfg.Headers = make(map[string]string)
		for k, v := range headersRaw {
			if s, ok := v.(string); ok {
				cfg.Headers[k] = s
			}
		}
	}
	
	return cfg, nil
}
