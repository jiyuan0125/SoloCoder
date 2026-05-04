package main

import (
	"os"
	"warranty-claim/internal/client"
)

const defaultServerURL = "http://localhost:8080"

func main() {
	serverURL := os.Getenv("WARRANTY_SERVER_URL")
	if serverURL == "" {
		serverURL = defaultServerURL
	}

	userID := os.Getenv("USER_ID")
	if userID == "" {
		userID = "default_user"
	}

	apiClient := client.NewAPIClient(serverURL, userID)
	cli := client.NewCLI(apiClient)
	cli.Run()
}
