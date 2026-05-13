package main

import (
	"fmt"
	"os"
	"path/filepath"

	"credits/cmd"
)

func main() {
	dataDir := os.Getenv("TRAINING_CREDITS_DATA_DIR")
	if dataDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取用户主目录失败: %v\n", err)
			os.Exit(1)
		}
		dataDir = filepath.Join(homeDir, ".training-credits")
	}

	cmd.Execute(dataDir)
}
