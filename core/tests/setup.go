// Package tests provides test utilities and configurations for the Bifrost system.
// It includes test implementations of schemas, mock objects, and helper functions
// for testing the Bifrost functionality with various AI providers.
package tests

import (
	"log"

	bifrost "github.com/maximhq/bifrost/core"
	schemas "github.com/maximhq/bifrost/core/schemas"

	"github.com/joho/godotenv"
)

// loadEnv loads environment variables from a .env file into the process environment.
// It uses the godotenv package to load variables and fails if the .env file cannot be loaded.
//
// Environment Variables:
//   - .env file: Contains configuration values for the test environment
//
// Returns:
//   - None, but will log.Fatal if the .env file cannot be loaded
func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}
}

// getBifrost initializes and returns a Bifrost instance for testing.
// It sets up the test account, plugin, and logger configuration.
//
// Environment Variables:
//   - Uses environment variables loaded by loadEnv()
//
// Returns:
//   - *bifrost.Bifrost: A configured Bifrost instance ready for testing
//   - error: Any error that occurred during Bifrost initialization
//
// The function:
//  1. Loads environment variables
//  2. Creates a test account instance
//  3. Initializes a plugin for request tracing
//  4. Configures Bifrost with the account, plugin, and default logger
func getBifrost() (*bifrost.Bifrost, error) {
	loadEnv()

	account := BaseAccount{}

	// Initialize Bifrost
	b, err := bifrost.Init(schemas.BifrostConfig{
		Account: &account,
		Plugins: nil,
		Logger:  bifrost.NewDefaultLogger(schemas.LogLevelDebug),
	})
	if err != nil {
		return nil, err
	}

	return b, nil
}
