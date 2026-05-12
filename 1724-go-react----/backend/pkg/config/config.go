package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	ServerPort string
	DBPath     string
}

func Load(portFlag string) *Config {
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("DB_PATH", "./teaching_eval.db")

	viper.AutomaticEnv()

	if portFlag != "" {
		viper.Set("PORT", portFlag)
	}

	port := viper.GetString("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		ServerPort: port,
		DBPath:     viper.GetString("DB_PATH"),
	}
}

func GetPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return "8080"
}

func GetAddr(port string) string {
	return fmt.Sprintf(":%s", port)
}
