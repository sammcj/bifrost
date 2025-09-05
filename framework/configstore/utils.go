package configstore

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/maximhq/bifrost/core/schemas"
)

// deepCopy creates a deep copy of a given type
func deepCopy[T any](in T) (T, error) {
	var out T
	b, err := json.Marshal(in)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(b, &out)
	return out, err
}

// substituteEnvVars replaces resolved environment variable values with their original env.VAR_NAME references
func substituteEnvVars(config *ProviderConfig, provider schemas.ModelProvider, envKeys map[string][]EnvKeyInfo) {
	// Create a map for quick lookup of env vars by provider and key ID
	envVarMap := make(map[string]string) // key: "provider.keyID.field" -> env var name

	for envVar, keyInfos := range envKeys {
		for _, keyInfo := range keyInfos {
			if keyInfo.Provider == provider {
				// For API keys
				if keyInfo.KeyType == "api_key" {
					envVarMap[fmt.Sprintf("%s.%s.value", provider, keyInfo.KeyID)] = envVar
				}
				// For Azure config
				if keyInfo.KeyType == "azure_config" {
					field := strings.TrimPrefix(keyInfo.ConfigPath, fmt.Sprintf("providers.%s.keys[%s].azure_key_config.", provider, keyInfo.KeyID))
					envVarMap[fmt.Sprintf("%s.%s.azure.%s", provider, keyInfo.KeyID, field)] = envVar
				}
				// For Vertex config
				if keyInfo.KeyType == "vertex_config" {
					field := strings.TrimPrefix(keyInfo.ConfigPath, fmt.Sprintf("providers.%s.keys[%s].vertex_key_config.", provider, keyInfo.KeyID))
					envVarMap[fmt.Sprintf("%s.%s.vertex.%s", provider, keyInfo.KeyID, field)] = envVar
				}
				// For Bedrock config
				if keyInfo.KeyType == "bedrock_config" {
					field := strings.TrimPrefix(keyInfo.ConfigPath, fmt.Sprintf("providers.%s.keys[%s].bedrock_key_config.", provider, keyInfo.KeyID))
					envVarMap[fmt.Sprintf("%s.%s.bedrock.%s", provider, keyInfo.KeyID, field)] = envVar
				}
			}
		}
	}

	// Substitute values in keys
	for i, key := range config.Keys {
		keyPrefix := fmt.Sprintf("%s.%s", provider, key.ID)

		// Substitute API key value
		if envVar, exists := envVarMap[fmt.Sprintf("%s.value", keyPrefix)]; exists {
			config.Keys[i].Value = fmt.Sprintf("env.%s", envVar)
		}

		// Substitute Azure config
		if key.AzureKeyConfig != nil {
			if envVar, exists := envVarMap[fmt.Sprintf("%s.azure.endpoint", keyPrefix)]; exists {
				config.Keys[i].AzureKeyConfig.Endpoint = fmt.Sprintf("env.%s", envVar)
			}
			if envVar, exists := envVarMap[fmt.Sprintf("%s.azure.api_version", keyPrefix)]; exists {
				apiVersion := fmt.Sprintf("env.%s", envVar)
				config.Keys[i].AzureKeyConfig.APIVersion = &apiVersion
			}
		}

		// Substitute Vertex config
		if key.VertexKeyConfig != nil {
			if envVar, exists := envVarMap[fmt.Sprintf("%s.vertex.project_id", keyPrefix)]; exists {
				config.Keys[i].VertexKeyConfig.ProjectID = fmt.Sprintf("env.%s", envVar)
			}
			if envVar, exists := envVarMap[fmt.Sprintf("%s.vertex.region", keyPrefix)]; exists {
				config.Keys[i].VertexKeyConfig.Region = fmt.Sprintf("env.%s", envVar)
			}
			if envVar, exists := envVarMap[fmt.Sprintf("%s.vertex.auth_credentials", keyPrefix)]; exists {
				config.Keys[i].VertexKeyConfig.AuthCredentials = fmt.Sprintf("env.%s", envVar)
			}
		}

		// Substitute Bedrock config
		if key.BedrockKeyConfig != nil {
			if envVar, exists := envVarMap[fmt.Sprintf("%s.bedrock.access_key", keyPrefix)]; exists {
				config.Keys[i].BedrockKeyConfig.AccessKey = fmt.Sprintf("env.%s", envVar)
			}
			if envVar, exists := envVarMap[fmt.Sprintf("%s.bedrock.secret_key", keyPrefix)]; exists {
				config.Keys[i].BedrockKeyConfig.SecretKey = fmt.Sprintf("env.%s", envVar)
			}
			if envVar, exists := envVarMap[fmt.Sprintf("%s.bedrock.session_token", keyPrefix)]; exists {
				config.Keys[i].BedrockKeyConfig.SessionToken = &[]string{fmt.Sprintf("env.%s", envVar)}[0]
			}
			if envVar, exists := envVarMap[fmt.Sprintf("%s.bedrock.region", keyPrefix)]; exists {
				config.Keys[i].BedrockKeyConfig.Region = &[]string{fmt.Sprintf("env.%s", envVar)}[0]
			}
			if envVar, exists := envVarMap[fmt.Sprintf("%s.bedrock.arn", keyPrefix)]; exists {
				config.Keys[i].BedrockKeyConfig.ARN = &[]string{fmt.Sprintf("env.%s", envVar)}[0]
			}
		}
	}
}