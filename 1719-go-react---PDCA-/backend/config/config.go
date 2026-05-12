package config

type Config struct {
	Port         int
	DatabasePath string
}

func Load() *Config {
	return &Config{
		Port:         8300,
		DatabasePath: "medical_quality.db",
	}
}
