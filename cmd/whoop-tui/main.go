package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/jvaikath/whoop-tui/internal/api"
	"github.com/jvaikath/whoop-tui/internal/auth"
	"github.com/jvaikath/whoop-tui/internal/tui"
)

func main() {
	godotenv.Load()

	clientID := os.Getenv("WHOOP_CLIENT_ID")
	clientSecret := os.Getenv("WHOOP_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		fmt.Fprintln(os.Stderr, "Set WHOOP_CLIENT_ID and WHOOP_CLIENT_SECRET environment variables.")
		fmt.Fprintln(os.Stderr, "Get these from https://developer.whoop.com")
		os.Exit(1)
	}

	httpClient, err := auth.GetClient(clientID, clientSecret)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Authentication failed: %v\n", err)
		os.Exit(1)
	}

	client := api.NewClient(httpClient)

	if err := tui.Run(client); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
