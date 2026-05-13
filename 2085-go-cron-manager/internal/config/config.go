package config

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v3"
)

type TaskConfig struct {
	Name        string `yaml:"name"`
	Cron        string `yaml:"cron"`
	Command     string `yaml:"command"`
	Description string `yaml:"description"`
	Retries     int    `yaml:"retries"`
	RetryDelay  int    `yaml:"retry_delay"`
	Paused      bool   `yaml:"paused"`
}

type Config struct {
	Tasks []*TaskConfig `yaml:"tasks"`
}

const DefaultConfigFile = "crontab.yaml"

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}

	var cfg Config
	dec := yaml.NewDecoder(newByteReader(data))
	dec.KnownFields(true)
	err = dec.Decode(&cfg)
	if err != nil {
		if yerr, ok := err.(*yaml.TypeError); ok {
			return nil, fmt.Errorf("YAML 格式错误: %v", yerr)
		}
		return nil, fmt.Errorf("YAML 解析错误: %v", err)
	}

	return &cfg, nil
}

type byteReader struct {
	buf *bytes.Buffer
}

func newByteReader(data []byte) *byteReader {
	return &byteReader{buf: bytes.NewBuffer(data)}
}

func (r *byteReader) Read(p []byte) (n int, err error) {
	return r.buf.Read(p)
}

func (r *byteReader) ReadByte() (byte, error) {
	b, err := r.buf.ReadByte()
	if err != nil {
		return 0, io.EOF
	}
	return b, nil
}

func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func ValidateCron(expr string) error {
	p := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := p.Parse(expr)
	return err
}
