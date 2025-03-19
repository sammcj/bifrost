package tests

import (
	"bifrost"
	"bifrost/interfaces"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}
}

func getBifrost() (*bifrost.Bifrost, error) {
	loadEnv()

	account := BaseAccount{}
	if err := account.Init(
		ProviderConfig{
			"openai": {
				Keys: []interfaces.Key{
					{Value: os.Getenv("OPEN_AI_API_KEY"), Weight: 1.0, Models: []string{"gpt-4o-mini"}},
				},
				ConcurrencyConfig: interfaces.ConcurrencyAndBufferSize{
					Concurrency: 3,
					BufferSize:  10,
				},
			},
			"anthropic": {
				Keys: []interfaces.Key{
					{Value: os.Getenv("ANTHROPIC_API_KEY"), Weight: 1.0, Models: []string{"claude-3-7-sonnet-20250219", "claude-2.1"}},
				},
				ConcurrencyConfig: interfaces.ConcurrencyAndBufferSize{
					Concurrency: 3,
					BufferSize:  10,
				},
			},
		},
	); err != nil {
		log.Fatal("Error initializing account:", err)
		return nil, err
	}

	bifrost, err := bifrost.Init(&account)
	if err != nil {
		log.Fatal("Error initializing bifrost:", err)
		return nil, err
	}

	return bifrost, nil
}
