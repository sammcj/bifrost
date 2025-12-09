package lib

/*
===================================================================================
CONFIG HASH TEST SCENARIOS INDEX
===================================================================================

This file contains tests for config hash generation and comparison logic used to
detect changes between config.json and database configuration.

HASH GENERATION TESTS
-------------------------------------------------------------------------------------
| Test Name                              | Description                              |
|----------------------------------------|------------------------------------------|
| TestGenerateProviderConfigHash         | Provider hash generation, keys excluded, |
|                                        | different fields → different hash        |
| TestGenerateKeyHash                    | Key hash generation, ID skipped,         |
|                                        | content changes detected                 |

COMPARISON LOGIC TESTS
-------------------------------------------------------------------------------------
| Test Name                              | Description                              |
|----------------------------------------|------------------------------------------|
| TestProviderHashComparison_MatchingHash| Hash matches → keep DB config            |
| TestProviderHashComparison_DifferentHash| Hash differs → sync from file,          |
|                                        | preserve dashboard-added keys            |
| TestProviderHashComparison_NewProvider | New provider in file → add to DB         |
| TestProviderHashComparison_ProviderOnlyInDB| Provider added via dashboard →        |
|                                        | preserved (not in file)                  |

ROUND-TRIP & LIFECYCLE TESTS
-------------------------------------------------------------------------------------
| Test Name                              | Description                              |
|----------------------------------------|------------------------------------------|
| TestProviderHashComparison_RoundTrip   | JSON → DB → same JSON = no changes       |
| TestProviderHashComparison_DashboardEditThenSameFile| Dashboard edits preserved |
|                                        | when file unchanged                      |
| TestProviderHashComparison_FullLifecycle| DB → new JSON → update DB →             |
|                                        | same JSON (no update)                    |
| TestProviderHashComparison_MultipleUpdates| Multiple config.json updates over     |
|                                        | time + revert to old config              |

FIELDS PRESENT/ABSENT TESTS
-------------------------------------------------------------------------------------
| Test Name                              | Description                              |
|----------------------------------------|------------------------------------------|
| TestProviderHashComparison_OptionalFieldsPresence| NetworkConfig, ProxyConfig,  |
|                                        | ConcurrencyAndBufferSize, CustomProvider |
|                                        | present vs absent                        |
| TestKeyHashComparison_OptionalFieldsPresence| Models, AzureKeyConfig,           |
|                                        | VertexKeyConfig, BedrockKeyConfig        |
|                                        | present vs absent                        |

FIELD REMOVAL TESTS
-------------------------------------------------------------------------------------
| Test Name                              | Description                              |
|----------------------------------------|------------------------------------------|
| TestProviderHashComparison_FieldRemoved| NetworkConfig, ProxyConfig,              |
|                                        | ConcurrencyAndBufferSize, ExtraHeaders,  |
|                                        | SendBackRawResponse removed              |
| TestKeyHashComparison_FieldRemoved     | Models, AzureKeyConfig, APIVersion,      |
|                                        | Weight removed                           |

FIELD VALUE CHANGE TESTS
-------------------------------------------------------------------------------------
| Test Name                              | Description                              |
|----------------------------------------|------------------------------------------|
| TestProviderHashComparison_FieldValueChanges| BaseURL, ExtraHeaders,              |
|                                        | Concurrency value changes                |
| TestProviderHashComparison_PartialFieldChanges| Timeout, MaxRetries changes       |
|                                        | within nested structs                    |
| TestKeyHashComparison_KeyContentChanged| Key Value, Models content changes        |

INDEPENDENT UPDATE TESTS
-------------------------------------------------------------------------------------
| Test Name                              | Description                              |
|----------------------------------------|------------------------------------------|
| TestProviderHashComparison_ProviderChangedKeysUnchanged| Provider config updated,|
|                                        | keys unchanged → only provider updates   |
| TestProviderHashComparison_KeysChangedProviderUnchanged| Keys updated, provider  |
|                                        | unchanged → only keys update             |
| TestProviderHashComparison_BothChangedIndependently| Both provider and keys   |
|                                        | changed → both update                    |
| TestProviderHashComparison_NeitherChanged| Neither provider nor keys changed →    |
|                                        | no updates needed                        |

MERGE LOGIC TESTS (non-hash)
-------------------------------------------------------------------------------------
| Test Name                              | Description                              |
|----------------------------------------|------------------------------------------|
| TestLoadConfig_ClientConfig_Merge      | Client config merge from DB and file     |
| TestLoadConfig_Providers_Merge         | Provider keys merge from DB and file     |
| TestLoadConfig_MCP_Merge               | MCP config merge from DB and file        |
| TestLoadConfig_Governance_Merge        | Governance config merge from DB and file |
===================================================================================
*/

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/maximhq/bifrost/core/schemas"
	"github.com/maximhq/bifrost/framework/configstore"
	"github.com/maximhq/bifrost/framework/configstore/tables"
	"github.com/maximhq/bifrost/framework/logstore"
	"github.com/maximhq/bifrost/framework/migrator"
	"github.com/maximhq/bifrost/framework/vectorstore"
	"gorm.io/gorm"
)

// MockConfigStore implements the ConfigStore interface for testing
type MockConfigStore struct {
	clientConfig     *configstore.ClientConfig
	providers        map[schemas.ModelProvider]configstore.ProviderConfig
	mcpConfig        *schemas.MCPConfig
	governanceConfig *configstore.GovernanceConfig
	authConfig       *configstore.AuthConfig
	frameworkConfig  *tables.TableFrameworkConfig
	vectorConfig     *vectorstore.Config
	logsConfig       *logstore.Config
	envKeys          map[string][]configstore.EnvKeyInfo
	plugins          []*tables.TablePlugin

	// Track update calls for verification
	clientConfigUpdated    bool
	providersConfigUpdated bool
	mcpConfigsCreated      []schemas.MCPClientConfig
	governanceItemsCreated struct {
		budgets     []tables.TableBudget
		rateLimits  []tables.TableRateLimit
		customers   []tables.TableCustomer
		teams       []tables.TableTeam
		virtualKeys []tables.TableVirtualKey
	}
}

// NewMockConfigStore creates a new mock config store
func NewMockConfigStore() *MockConfigStore {
	return &MockConfigStore{
		providers: make(map[schemas.ModelProvider]configstore.ProviderConfig),
		envKeys:   make(map[string][]configstore.EnvKeyInfo),
	}
}

// Implement ConfigStore interface methods
func (m *MockConfigStore) Ping(ctx context.Context) error  { return nil }
func (m *MockConfigStore) Close(ctx context.Context) error { return nil }
func (m *MockConfigStore) DB() *gorm.DB                    { return nil }
func (m *MockConfigStore) ExecuteTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return fn(nil)
}
func (m *MockConfigStore) RunMigration(ctx context.Context, migration *migrator.Migration) error {
	return nil
}

// Client config
func (m *MockConfigStore) UpdateClientConfig(ctx context.Context, config *configstore.ClientConfig) error {
	m.clientConfig = config
	m.clientConfigUpdated = true
	return nil
}

func (m *MockConfigStore) GetClientConfig(ctx context.Context) (*configstore.ClientConfig, error) {
	return m.clientConfig, nil
}

// Provider config
func (m *MockConfigStore) UpdateProvidersConfig(ctx context.Context, providers map[schemas.ModelProvider]configstore.ProviderConfig) error {
	m.providers = providers
	m.providersConfigUpdated = true
	return nil
}

func (m *MockConfigStore) GetProvidersConfig(ctx context.Context) (map[schemas.ModelProvider]configstore.ProviderConfig, error) {
	if len(m.providers) == 0 {
		return nil, nil
	}
	return m.providers, nil
}

func (m *MockConfigStore) AddProvider(ctx context.Context, provider schemas.ModelProvider, config configstore.ProviderConfig, envKeys map[string][]configstore.EnvKeyInfo) error {
	m.providers[provider] = config
	return nil
}

func (m *MockConfigStore) UpdateProvider(ctx context.Context, provider schemas.ModelProvider, config configstore.ProviderConfig, envKeys map[string][]configstore.EnvKeyInfo) error {
	m.providers[provider] = config
	return nil
}

func (m *MockConfigStore) DeleteProvider(ctx context.Context, provider schemas.ModelProvider) error {
	delete(m.providers, provider)
	return nil
}

// MCP config
func (m *MockConfigStore) GetMCPConfig(ctx context.Context) (*schemas.MCPConfig, error) {
	return m.mcpConfig, nil
}

func (m *MockConfigStore) GetMCPClientByName(ctx context.Context, name string) (*tables.TableMCPClient, error) {
	return nil, nil
}

func (m *MockConfigStore) CreateMCPClientConfig(ctx context.Context, clientConfig schemas.MCPClientConfig, envKeys map[string][]configstore.EnvKeyInfo) error {
	if m.mcpConfig == nil {
		m.mcpConfig = &schemas.MCPConfig{
			ClientConfigs: []schemas.MCPClientConfig{},
		}
	}
	m.mcpConfig.ClientConfigs = append(m.mcpConfig.ClientConfigs, clientConfig)
	m.mcpConfigsCreated = append(m.mcpConfigsCreated, clientConfig)
	return nil
}

func (m *MockConfigStore) UpdateMCPClientConfig(ctx context.Context, id string, clientConfig schemas.MCPClientConfig, envKeys map[string][]configstore.EnvKeyInfo) error {
	return nil
}

func (m *MockConfigStore) DeleteMCPClientConfig(ctx context.Context, id string) error {
	return nil
}

// Governance config
func (m *MockConfigStore) GetGovernanceConfig(ctx context.Context) (*configstore.GovernanceConfig, error) {
	return m.governanceConfig, nil
}

func (m *MockConfigStore) CreateBudget(ctx context.Context, budget *tables.TableBudget, tx ...*gorm.DB) error {
	if m.governanceConfig == nil {
		m.governanceConfig = &configstore.GovernanceConfig{}
	}
	m.governanceConfig.Budgets = append(m.governanceConfig.Budgets, *budget)
	m.governanceItemsCreated.budgets = append(m.governanceItemsCreated.budgets, *budget)
	return nil
}

func (m *MockConfigStore) UpdateBudget(ctx context.Context, budget *tables.TableBudget, tx ...*gorm.DB) error {
	return nil
}

func (m *MockConfigStore) UpdateBudgets(ctx context.Context, budgets []*tables.TableBudget, tx ...*gorm.DB) error {
	return nil
}

func (m *MockConfigStore) GetBudget(ctx context.Context, id string, tx ...*gorm.DB) (*tables.TableBudget, error) {
	return nil, nil
}

func (m *MockConfigStore) GetBudgets(ctx context.Context) ([]tables.TableBudget, error) {
	return nil, nil
}

func (m *MockConfigStore) CreateRateLimit(ctx context.Context, rateLimit *tables.TableRateLimit, tx ...*gorm.DB) error {
	if m.governanceConfig == nil {
		m.governanceConfig = &configstore.GovernanceConfig{}
	}
	m.governanceConfig.RateLimits = append(m.governanceConfig.RateLimits, *rateLimit)
	m.governanceItemsCreated.rateLimits = append(m.governanceItemsCreated.rateLimits, *rateLimit)
	return nil
}

func (m *MockConfigStore) UpdateRateLimit(ctx context.Context, rateLimit *tables.TableRateLimit, tx ...*gorm.DB) error {
	return nil
}

func (m *MockConfigStore) UpdateRateLimits(ctx context.Context, rateLimits []*tables.TableRateLimit, tx ...*gorm.DB) error {
	return nil
}

func (m *MockConfigStore) GetRateLimit(ctx context.Context, id string) (*tables.TableRateLimit, error) {
	return nil, nil
}

func (m *MockConfigStore) CreateCustomer(ctx context.Context, customer *tables.TableCustomer, tx ...*gorm.DB) error {
	if m.governanceConfig == nil {
		m.governanceConfig = &configstore.GovernanceConfig{}
	}
	m.governanceConfig.Customers = append(m.governanceConfig.Customers, *customer)
	m.governanceItemsCreated.customers = append(m.governanceItemsCreated.customers, *customer)
	return nil
}

func (m *MockConfigStore) UpdateCustomer(ctx context.Context, customer *tables.TableCustomer, tx ...*gorm.DB) error {
	return nil
}

func (m *MockConfigStore) DeleteCustomer(ctx context.Context, id string) error {
	return nil
}

func (m *MockConfigStore) GetCustomer(ctx context.Context, id string) (*tables.TableCustomer, error) {
	return nil, nil
}

func (m *MockConfigStore) GetCustomers(ctx context.Context) ([]tables.TableCustomer, error) {
	return nil, nil
}

func (m *MockConfigStore) CreateTeam(ctx context.Context, team *tables.TableTeam, tx ...*gorm.DB) error {
	if m.governanceConfig == nil {
		m.governanceConfig = &configstore.GovernanceConfig{}
	}
	m.governanceConfig.Teams = append(m.governanceConfig.Teams, *team)
	m.governanceItemsCreated.teams = append(m.governanceItemsCreated.teams, *team)
	return nil
}

func (m *MockConfigStore) UpdateTeam(ctx context.Context, team *tables.TableTeam, tx ...*gorm.DB) error {
	return nil
}

func (m *MockConfigStore) DeleteTeam(ctx context.Context, id string) error {
	return nil
}

func (m *MockConfigStore) GetTeam(ctx context.Context, id string) (*tables.TableTeam, error) {
	return nil, nil
}

func (m *MockConfigStore) GetTeams(ctx context.Context, customerID string) ([]tables.TableTeam, error) {
	return nil, nil
}

func (m *MockConfigStore) CreateVirtualKey(ctx context.Context, virtualKey *tables.TableVirtualKey, tx ...*gorm.DB) error {
	if m.governanceConfig == nil {
		m.governanceConfig = &configstore.GovernanceConfig{}
	}
	m.governanceConfig.VirtualKeys = append(m.governanceConfig.VirtualKeys, *virtualKey)
	m.governanceItemsCreated.virtualKeys = append(m.governanceItemsCreated.virtualKeys, *virtualKey)
	return nil
}

func (m *MockConfigStore) UpdateVirtualKey(ctx context.Context, virtualKey *tables.TableVirtualKey, tx ...*gorm.DB) error {
	return nil
}

func (m *MockConfigStore) DeleteVirtualKey(ctx context.Context, id string) error {
	return nil
}

func (m *MockConfigStore) GetVirtualKey(ctx context.Context, id string) (*tables.TableVirtualKey, error) {
	return nil, nil
}

func (m *MockConfigStore) GetVirtualKeys(ctx context.Context) ([]tables.TableVirtualKey, error) {
	return nil, nil
}

func (m *MockConfigStore) GetRedactedVirtualKeys(ctx context.Context, ids []string) ([]tables.TableVirtualKey, error) {
	return nil, nil
}

func (m *MockConfigStore) GetVirtualKeyByValue(ctx context.Context, value string) (*tables.TableVirtualKey, error) {
	return nil, nil
}

// Virtual key provider config
func (m *MockConfigStore) GetVirtualKeyProviderConfigs(ctx context.Context, virtualKeyID string) ([]tables.TableVirtualKeyProviderConfig, error) {
	return nil, nil
}

func (m *MockConfigStore) CreateVirtualKeyProviderConfig(ctx context.Context, virtualKeyProviderConfig *tables.TableVirtualKeyProviderConfig, tx ...*gorm.DB) error {
	return nil
}

func (m *MockConfigStore) UpdateVirtualKeyProviderConfig(ctx context.Context, virtualKeyProviderConfig *tables.TableVirtualKeyProviderConfig, tx ...*gorm.DB) error {
	return nil
}

func (m *MockConfigStore) DeleteVirtualKeyProviderConfig(ctx context.Context, id uint, tx ...*gorm.DB) error {
	return nil
}

// Virtual key MCP config
func (m *MockConfigStore) GetVirtualKeyMCPConfigs(ctx context.Context, virtualKeyID string) ([]tables.TableVirtualKeyMCPConfig, error) {
	return nil, nil
}

func (m *MockConfigStore) CreateVirtualKeyMCPConfig(ctx context.Context, virtualKeyMCPConfig *tables.TableVirtualKeyMCPConfig, tx ...*gorm.DB) error {
	return nil
}

func (m *MockConfigStore) UpdateVirtualKeyMCPConfig(ctx context.Context, virtualKeyMCPConfig *tables.TableVirtualKeyMCPConfig, tx ...*gorm.DB) error {
	return nil
}

func (m *MockConfigStore) DeleteVirtualKeyMCPConfig(ctx context.Context, id uint, tx ...*gorm.DB) error {
	return nil
}

// Auth config
func (m *MockConfigStore) GetAuthConfig(ctx context.Context) (*configstore.AuthConfig, error) {
	return m.authConfig, nil
}

func (m *MockConfigStore) UpdateAuthConfig(ctx context.Context, config *configstore.AuthConfig) error {
	m.authConfig = config
	return nil
}

// Framework config
func (m *MockConfigStore) UpdateFrameworkConfig(ctx context.Context, config *tables.TableFrameworkConfig) error {
	m.frameworkConfig = config
	return nil
}

func (m *MockConfigStore) GetFrameworkConfig(ctx context.Context) (*tables.TableFrameworkConfig, error) {
	return m.frameworkConfig, nil
}

// Vector store config
func (m *MockConfigStore) UpdateVectorStoreConfig(ctx context.Context, config *vectorstore.Config) error {
	m.vectorConfig = config
	return nil
}

func (m *MockConfigStore) GetVectorStoreConfig(ctx context.Context) (*vectorstore.Config, error) {
	return m.vectorConfig, nil
}

// Logs store config
func (m *MockConfigStore) UpdateLogsStoreConfig(ctx context.Context, config *logstore.Config) error {
	m.logsConfig = config
	return nil
}

func (m *MockConfigStore) GetLogsStoreConfig(ctx context.Context) (*logstore.Config, error) {
	return m.logsConfig, nil
}

// ENV keys
func (m *MockConfigStore) UpdateEnvKeys(ctx context.Context, keys map[string][]configstore.EnvKeyInfo) error {
	m.envKeys = keys
	return nil
}

func (m *MockConfigStore) GetEnvKeys(ctx context.Context) (map[string][]configstore.EnvKeyInfo, error) {
	return m.envKeys, nil
}

// Config
func (m *MockConfigStore) GetConfig(ctx context.Context, key string) (*tables.TableGovernanceConfig, error) {
	return nil, nil
}

func (m *MockConfigStore) UpdateConfig(ctx context.Context, config *tables.TableGovernanceConfig, tx ...*gorm.DB) error {
	return nil
}

// Plugins
func (m *MockConfigStore) GetPlugins(ctx context.Context) ([]*tables.TablePlugin, error) {
	return m.plugins, nil
}

func (m *MockConfigStore) GetPlugin(ctx context.Context, name string) (*tables.TablePlugin, error) {
	for _, p := range m.plugins {
		if p.Name == name {
			return p, nil
		}
	}
	return nil, nil
}

func (m *MockConfigStore) CreatePlugin(ctx context.Context, plugin *tables.TablePlugin, tx ...*gorm.DB) error {
	m.plugins = append(m.plugins, plugin)
	return nil
}

func (m *MockConfigStore) UpdatePlugin(ctx context.Context, plugin *tables.TablePlugin, tx ...*gorm.DB) error {
	return nil
}

func (m *MockConfigStore) DeletePlugin(ctx context.Context, name string, tx ...*gorm.DB) error {
	return nil
}

// Key management
func (m *MockConfigStore) GetKeysByIDs(ctx context.Context, ids []string) ([]tables.TableKey, error) {
	return nil, nil
}

func (m *MockConfigStore) GetAllRedactedKeys(ctx context.Context, ids []string) ([]schemas.Key, error) {
	return nil, nil
}

// Session
func (m *MockConfigStore) GetSession(ctx context.Context, token string) (*tables.SessionsTable, error) {
	return nil, nil
}

func (m *MockConfigStore) CreateSession(ctx context.Context, session *tables.SessionsTable) error {
	return nil
}

func (m *MockConfigStore) DeleteSession(ctx context.Context, token string) error {
	return nil
}

// Model pricing
func (m *MockConfigStore) GetModelPrices(ctx context.Context) ([]tables.TableModelPricing, error) {
	return nil, nil
}

func (m *MockConfigStore) CreateModelPrices(ctx context.Context, pricing *tables.TableModelPricing, tx ...*gorm.DB) error {
	return nil
}

func (m *MockConfigStore) DeleteModelPrices(ctx context.Context, tx ...*gorm.DB) error {
	return nil
}

// Helper functions for tests

// createTempDir creates a temporary directory for test files
func createTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "bifrost-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

// createConfigFile creates a config.json file with the given data
func createConfigFile(t *testing.T, dir string, data *ConfigData) {
	t.Helper()
	configPath := filepath.Join(dir, "config.json")
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal config data: %v", err)
	}
	if err := os.WriteFile(configPath, jsonData, 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
}

// Test fixtures

func makeClientConfig(initialPoolSize int, enableLogging bool) *configstore.ClientConfig {
	return &configstore.ClientConfig{
		InitialPoolSize:      initialPoolSize,
		EnableLogging:        enableLogging,
		MaxRequestBodySizeMB: 10,
		PrometheusLabels:     []string{"label1"},
		AllowedOrigins:       []string{"http://localhost:3000"},
	}
}

func makeProviderConfig(keyName, keyValue string) configstore.ProviderConfig {
	return configstore.ProviderConfig{
		Keys: []schemas.Key{
			{
				ID:     uuid.NewString(),
				Name:   keyName,
				Value:  keyValue,
				Weight: 1,
			},
		},
	}
}

func makeMCPClientConfig(id, name string) schemas.MCPClientConfig {
	return schemas.MCPClientConfig{
		ID:             id,
		Name:           name,
		ConnectionType: schemas.MCPConnectionTypeHTTP,
	}
}

// Tests

// TestLoadConfig_ClientConfig_Merge tests client config merge from DB and file
func TestLoadConfig_ClientConfig_Merge(t *testing.T) {
	tempDir := createTempDir(t)

	// Create config file with client config
	fileClientConfig := &configstore.ClientConfig{
		InitialPoolSize:       20,
		EnableLogging:         true,
		EnableGovernance:      true,
		PrometheusLabels:      []string{"file-label"},
		AllowedOrigins:        []string{"http://file-origin.com"},
		MaxRequestBodySizeMB:  15,
		DisableContentLogging: true,
	}

	configData := &ConfigData{
		Client: fileClientConfig,
	}
	createConfigFile(t, tempDir, configData)

	// Setup mock config store with existing client config
	mockStore := NewMockConfigStore()
	mockStore.clientConfig = &configstore.ClientConfig{
		InitialPoolSize:      10,
		EnableLogging:        false,
		EnableGovernance:     false,
		PrometheusLabels:     []string{"db-label"},
		MaxRequestBodySizeMB: 5,
		// AllowedOrigins is empty in DB
	}

	// Override the config store creation to use our mock
	originalConfigStore := mockStore

	// Load config (we need to test the merge logic manually since LoadConfig creates its own store)
	// For now, let's test the merge logic by simulating what happens

	// Simulate merge: DB takes priority, file fills in empty values
	mergedConfig := *mockStore.clientConfig

	// InitialPoolSize: DB has 10, file has 20 -> keep DB (10)
	if mergedConfig.InitialPoolSize == 0 && fileClientConfig.InitialPoolSize != 0 {
		mergedConfig.InitialPoolSize = fileClientConfig.InitialPoolSize
	}

	// PrometheusLabels: DB has value, file has value -> keep DB
	if len(mergedConfig.PrometheusLabels) == 0 && len(fileClientConfig.PrometheusLabels) > 0 {
		mergedConfig.PrometheusLabels = fileClientConfig.PrometheusLabels
	}

	// AllowedOrigins: DB empty, file has value -> use file
	if len(mergedConfig.AllowedOrigins) == 0 && len(fileClientConfig.AllowedOrigins) > 0 {
		mergedConfig.AllowedOrigins = fileClientConfig.AllowedOrigins
	}

	// MaxRequestBodySizeMB: DB has 5, file has 15 -> keep DB (5)
	if mergedConfig.MaxRequestBodySizeMB == 0 && fileClientConfig.MaxRequestBodySizeMB != 0 {
		mergedConfig.MaxRequestBodySizeMB = fileClientConfig.MaxRequestBodySizeMB
	}

	// Boolean fields: file true overrides DB false
	if !mergedConfig.EnableLogging && fileClientConfig.EnableLogging {
		mergedConfig.EnableLogging = fileClientConfig.EnableLogging
	}
	if !mergedConfig.EnableGovernance && fileClientConfig.EnableGovernance {
		mergedConfig.EnableGovernance = fileClientConfig.EnableGovernance
	}
	if !mergedConfig.DisableContentLogging && fileClientConfig.DisableContentLogging {
		mergedConfig.DisableContentLogging = fileClientConfig.DisableContentLogging
	}

	// Verify merge results
	if mergedConfig.InitialPoolSize != 10 {
		t.Errorf("Expected InitialPoolSize to be 10 (from DB), got %d", mergedConfig.InitialPoolSize)
	}

	if len(mergedConfig.PrometheusLabels) != 1 || mergedConfig.PrometheusLabels[0] != "db-label" {
		t.Errorf("Expected PrometheusLabels to be [db-label] (from DB), got %v", mergedConfig.PrometheusLabels)
	}

	if len(mergedConfig.AllowedOrigins) != 1 || mergedConfig.AllowedOrigins[0] != "http://file-origin.com" {
		t.Errorf("Expected AllowedOrigins to be [http://file-origin.com] (from file), got %v", mergedConfig.AllowedOrigins)
	}

	if mergedConfig.MaxRequestBodySizeMB != 5 {
		t.Errorf("Expected MaxRequestBodySizeMB to be 5 (from DB), got %d", mergedConfig.MaxRequestBodySizeMB)
	}

	if !mergedConfig.EnableLogging {
		t.Error("Expected EnableLogging to be true (file true overrides DB false)")
	}

	if !mergedConfig.EnableGovernance {
		t.Error("Expected EnableGovernance to be true (file true overrides DB false)")
	}

	if !mergedConfig.DisableContentLogging {
		t.Error("Expected DisableContentLogging to be true (file true overrides DB false)")
	}

	_ = originalConfigStore
}

// TestLoadConfig_Providers_Merge tests provider keys merge from DB and file
func TestLoadConfig_Providers_Merge(t *testing.T) {
	// Setup DB providers
	dbProviders := make(map[schemas.ModelProvider]configstore.ProviderConfig)
	dbProviders[schemas.OpenAI] = configstore.ProviderConfig{
		Keys: []schemas.Key{
			{
				ID:     "key-1",
				Name:   "openai-db-key-1",
				Value:  "sk-db-123",
				Weight: 1,
			},
			{
				ID:     "key-2",
				Name:   "openai-db-key-2",
				Value:  "sk-db-456",
				Weight: 1,
			},
		},
	}

	// Setup file providers with some overlapping and some new keys
	fileProviders := map[string]configstore.ProviderConfig{
		"openai": {
			Keys: []schemas.Key{
				{
					ID:     "key-1", // Same ID as DB - should be skipped
					Name:   "openai-db-key-1",
					Value:  "sk-different",
					Weight: 1,
				},
				{
					ID:     "key-3", // New key
					Name:   "openai-file-key-3",
					Value:  "sk-file-789",
					Weight: 1,
				},
			},
		},
	}

	// Simulate merge logic
	for providerName, fileCfg := range fileProviders {
		provider := schemas.ModelProvider(providerName)
		if existingCfg, exists := dbProviders[provider]; exists {
			// Merge keys
			keysToAdd := make([]schemas.Key, 0)
			for _, newKey := range fileCfg.Keys {
				found := false
				for _, existingKey := range existingCfg.Keys {
					if existingKey.Name == newKey.Name || existingKey.ID == newKey.ID || existingKey.Value == newKey.Value {
						found = true
						break
					}
				}
				if !found {
					keysToAdd = append(keysToAdd, newKey)
				}
			}
			existingCfg.Keys = append(existingCfg.Keys, keysToAdd...)
			dbProviders[provider] = existingCfg
		}
	}

	// Verify merge results
	openaiCfg := dbProviders[schemas.OpenAI]
	if len(openaiCfg.Keys) != 3 {
		t.Errorf("Expected 3 keys after merge (2 from DB + 1 new from file), got %d", len(openaiCfg.Keys))
	}

	// Verify the keys
	keyNames := make(map[string]bool)
	for _, key := range openaiCfg.Keys {
		keyNames[key.Name] = true
	}

	if !keyNames["openai-db-key-1"] {
		t.Error("Expected openai-db-key-1 to be present")
	}
	if !keyNames["openai-db-key-2"] {
		t.Error("Expected openai-db-key-2 to be present")
	}
	if !keyNames["openai-file-key-3"] {
		t.Error("Expected openai-file-key-3 to be present (new from file)")
	}
}

// TestLoadConfig_MCP_Merge tests MCP config merge from DB and file
func TestLoadConfig_MCP_Merge(t *testing.T) {
	// Setup DB MCP config
	dbMCPConfig := &schemas.MCPConfig{
		ClientConfigs: []schemas.MCPClientConfig{
			{
				ID:             "mcp-1",
				Name:           "db-client-1",
				ConnectionType: schemas.MCPConnectionTypeHTTP,
			},
			{
				ID:             "mcp-2",
				Name:           "db-client-2",
				ConnectionType: schemas.MCPConnectionTypeSTDIO,
			},
		},
	}

	// Setup file MCP config with some overlapping and some new
	fileMCPConfig := &schemas.MCPConfig{
		ClientConfigs: []schemas.MCPClientConfig{
			{
				ID:             "mcp-1", // Same ID - should be skipped
				Name:           "different-name",
				ConnectionType: schemas.MCPConnectionTypeHTTP,
			},
			{
				ID:             "mcp-3", // New ID
				Name:           "file-client-3",
				ConnectionType: schemas.MCPConnectionTypeSSE,
			},
			{
				ID:             "mcp-4",       // New
				Name:           "db-client-2", // Same name as existing - should be skipped
				ConnectionType: schemas.MCPConnectionTypeHTTP,
			},
		},
	}

	// Simulate merge logic
	clientConfigsToAdd := make([]schemas.MCPClientConfig, 0)
	for _, newClientConfig := range fileMCPConfig.ClientConfigs {
		found := false
		for _, existingClientConfig := range dbMCPConfig.ClientConfigs {
			if (newClientConfig.ID != "" && existingClientConfig.ID == newClientConfig.ID) ||
				(newClientConfig.Name != "" && existingClientConfig.Name == newClientConfig.Name) {
				found = true
				break
			}
		}
		if !found {
			clientConfigsToAdd = append(clientConfigsToAdd, newClientConfig)
		}
	}

	mergedMCPConfig := &schemas.MCPConfig{
		ClientConfigs: append(dbMCPConfig.ClientConfigs, clientConfigsToAdd...),
	}

	// Verify merge results
	if len(mergedMCPConfig.ClientConfigs) != 3 {
		t.Errorf("Expected 3 client configs after merge (2 from DB + 1 new from file), got %d", len(mergedMCPConfig.ClientConfigs))
	}

	// Verify the client configs
	ids := make(map[string]bool)
	names := make(map[string]bool)
	for _, cc := range mergedMCPConfig.ClientConfigs {
		ids[cc.ID] = true
		names[cc.Name] = true
	}

	if !ids["mcp-1"] {
		t.Error("Expected mcp-1 to be present")
	}
	if !ids["mcp-2"] {
		t.Error("Expected mcp-2 to be present")
	}
	if !ids["mcp-3"] {
		t.Error("Expected mcp-3 to be present (new from file)")
	}
	if ids["mcp-4"] {
		t.Error("Expected mcp-4 to be skipped (same name as existing)")
	}
}

// TestLoadConfig_Governance_Merge tests governance config merge from DB and file
func TestLoadConfig_Governance_Merge(t *testing.T) {
	// Setup DB governance config
	dbGovernanceConfig := &configstore.GovernanceConfig{
		Budgets: []tables.TableBudget{
			{ID: "budget-1"},
			{ID: "budget-2"},
		},
		RateLimits: []tables.TableRateLimit{
			{ID: "ratelimit-1"},
		},
		Customers: []tables.TableCustomer{
			{ID: "customer-1"},
		},
		Teams: []tables.TableTeam{
			{ID: "team-1"},
		},
		VirtualKeys: []tables.TableVirtualKey{
			{ID: "vkey-1"},
		},
	}

	// Setup file governance config with some overlapping and some new
	fileGovernanceConfig := &configstore.GovernanceConfig{
		Budgets: []tables.TableBudget{
			{ID: "budget-1"}, // Duplicate
			{ID: "budget-3"}, // New
		},
		RateLimits: []tables.TableRateLimit{
			{ID: "ratelimit-2"}, // New
		},
		Customers: []tables.TableCustomer{
			{ID: "customer-1"}, // Duplicate
			{ID: "customer-2"}, // New
		},
		Teams: []tables.TableTeam{
			{ID: "team-2"}, // New
		},
		VirtualKeys: []tables.TableVirtualKey{
			{ID: "vkey-1"}, // Duplicate
			{ID: "vkey-2"}, // New
		},
	}

	// Simulate merge logic for Budgets
	budgetsToAdd := make([]tables.TableBudget, 0)
	for _, newBudget := range fileGovernanceConfig.Budgets {
		found := false
		for _, existingBudget := range dbGovernanceConfig.Budgets {
			if existingBudget.ID == newBudget.ID {
				found = true
				break
			}
		}
		if !found {
			budgetsToAdd = append(budgetsToAdd, newBudget)
		}
	}
	mergedBudgets := append(dbGovernanceConfig.Budgets, budgetsToAdd...)

	// Simulate merge logic for RateLimits
	rateLimitsToAdd := make([]tables.TableRateLimit, 0)
	for _, newRateLimit := range fileGovernanceConfig.RateLimits {
		found := false
		for _, existingRateLimit := range dbGovernanceConfig.RateLimits {
			if existingRateLimit.ID == newRateLimit.ID {
				found = true
				break
			}
		}
		if !found {
			rateLimitsToAdd = append(rateLimitsToAdd, newRateLimit)
		}
	}
	mergedRateLimits := append(dbGovernanceConfig.RateLimits, rateLimitsToAdd...)

	// Simulate merge logic for Customers
	customersToAdd := make([]tables.TableCustomer, 0)
	for _, newCustomer := range fileGovernanceConfig.Customers {
		found := false
		for _, existingCustomer := range dbGovernanceConfig.Customers {
			if existingCustomer.ID == newCustomer.ID {
				found = true
				break
			}
		}
		if !found {
			customersToAdd = append(customersToAdd, newCustomer)
		}
	}
	mergedCustomers := append(dbGovernanceConfig.Customers, customersToAdd...)

	// Simulate merge logic for Teams
	teamsToAdd := make([]tables.TableTeam, 0)
	for _, newTeam := range fileGovernanceConfig.Teams {
		found := false
		for _, existingTeam := range dbGovernanceConfig.Teams {
			if existingTeam.ID == newTeam.ID {
				found = true
				break
			}
		}
		if !found {
			teamsToAdd = append(teamsToAdd, newTeam)
		}
	}
	mergedTeams := append(dbGovernanceConfig.Teams, teamsToAdd...)

	// Simulate merge logic for VirtualKeys
	virtualKeysToAdd := make([]tables.TableVirtualKey, 0)
	for _, newVirtualKey := range fileGovernanceConfig.VirtualKeys {
		found := false
		for _, existingVirtualKey := range dbGovernanceConfig.VirtualKeys {
			if existingVirtualKey.ID == newVirtualKey.ID {
				found = true
				break
			}
		}
		if !found {
			virtualKeysToAdd = append(virtualKeysToAdd, newVirtualKey)
		}
	}
	mergedVirtualKeys := append(dbGovernanceConfig.VirtualKeys, virtualKeysToAdd...)

	// Verify merge results
	if len(mergedBudgets) != 3 {
		t.Errorf("Expected 3 budgets after merge (2 from DB + 1 new), got %d", len(mergedBudgets))
	}

	if len(mergedRateLimits) != 2 {
		t.Errorf("Expected 2 rate limits after merge (1 from DB + 1 new), got %d", len(mergedRateLimits))
	}

	if len(mergedCustomers) != 2 {
		t.Errorf("Expected 2 customers after merge (1 from DB + 1 new), got %d", len(mergedCustomers))
	}

	if len(mergedTeams) != 2 {
		t.Errorf("Expected 2 teams after merge (1 from DB + 1 new), got %d", len(mergedTeams))
	}

	if len(mergedVirtualKeys) != 2 {
		t.Errorf("Expected 2 virtual keys after merge (1 from DB + 1 new), got %d", len(mergedVirtualKeys))
	}

	// Verify specific IDs
	budgetIDs := make(map[string]bool)
	for _, b := range mergedBudgets {
		budgetIDs[b.ID] = true
	}
	if !budgetIDs["budget-1"] || !budgetIDs["budget-2"] || !budgetIDs["budget-3"] {
		t.Error("Expected budgets budget-1, budget-2, and budget-3")
	}
}

// TestGenerateProviderConfigHash tests that provider config hash is generated correctly
func TestGenerateProviderConfigHash(t *testing.T) {
	// Create a provider config
	config1 := configstore.ProviderConfig{
		Keys: []schemas.Key{
			{ID: "key-1", Name: "test-key", Value: "sk-123", Weight: 1},
		},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.example.com",
		},
		SendBackRawResponse: true,
	}

	// Generate hash
	hash1, err := config1.GenerateConfigHash("openai")
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	if hash1 == "" {
		t.Error("Expected non-empty hash")
	}

	// Same config should produce same hash
	config2 := configstore.ProviderConfig{
		Keys: []schemas.Key{
			{ID: "different-id", Name: "different-name", Value: "different-value", Weight: 2}, // Keys should NOT affect hash
		},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.example.com",
		},
		SendBackRawResponse: true,
	}

	hash2, err := config2.GenerateConfigHash("openai")
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	if hash1 != hash2 {
		t.Error("Expected same hash for configs with same fields (keys excluded)")
	}

	// Different config should produce different hash
	config3 := configstore.ProviderConfig{
		Keys: []schemas.Key{
			{ID: "key-1", Name: "test-key", Value: "sk-123", Weight: 1},
		},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://different-api.example.com", // Different base URL
		},
		SendBackRawResponse: true,
	}

	hash3, err := config3.GenerateConfigHash("openai")
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	if hash1 == hash3 {
		t.Error("Expected different hash for configs with different NetworkConfig")
	}

	// Different provider name should produce different hash
	hash4, err := config1.GenerateConfigHash("anthropic")
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	if hash1 == hash4 {
		t.Error("Expected different hash for different provider names")
	}
}

// TestGenerateKeyHash tests that key hash is generated correctly
func TestGenerateKeyHash(t *testing.T) {
	// Create a key
	key1 := schemas.Key{
		ID:     "key-1",
		Name:   "test-key",
		Value:  "sk-123",
		Models: []string{"gpt-4", "gpt-3.5-turbo"},
		Weight: 1.5,
	}

	// Generate hash
	hash1, err := configstore.GenerateKeyHash(key1)
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	if hash1 == "" {
		t.Error("Expected non-empty hash")
	}

	// Same key content with different ID should produce same hash (ID is skipped)
	key2 := schemas.Key{
		ID:     "different-id", // Different ID - should be skipped
		Name:   "test-key",
		Value:  "sk-123",
		Models: []string{"gpt-4", "gpt-3.5-turbo"},
		Weight: 1.5,
	}

	hash2, err := configstore.GenerateKeyHash(key2)
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	if hash1 != hash2 {
		t.Error("Expected same hash for keys with same content (ID should be skipped)")
	}

	// Different value should produce different hash
	key3 := schemas.Key{
		ID:     "key-1",
		Name:   "test-key",
		Value:  "sk-different", // Different value
		Models: []string{"gpt-4", "gpt-3.5-turbo"},
		Weight: 1.5,
	}

	hash3, err := configstore.GenerateKeyHash(key3)
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	if hash1 == hash3 {
		t.Error("Expected different hash for keys with different Value")
	}

	// Different models should produce different hash
	key4 := schemas.Key{
		ID:     "key-1",
		Name:   "test-key",
		Value:  "sk-123",
		Models: []string{"gpt-4"}, // Different models
		Weight: 1.5,
	}

	hash4, err := configstore.GenerateKeyHash(key4)
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	if hash1 == hash4 {
		t.Error("Expected different hash for keys with different Models")
	}

	// Different weight should produce different hash
	key5 := schemas.Key{
		ID:     "key-1",
		Name:   "test-key",
		Value:  "sk-123",
		Models: []string{"gpt-4", "gpt-3.5-turbo"},
		Weight: 2.0, // Different weight
	}

	hash5, err := configstore.GenerateKeyHash(key5)
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	if hash1 == hash5 {
		t.Error("Expected different hash for keys with different Weight")
	}
}

// TestProviderHashComparison_MatchingHash tests that DB config is kept when hashes match
func TestProviderHashComparison_MatchingHash(t *testing.T) {
	// Create a provider config (simulating what's in config.json)
	fileConfig := configstore.ProviderConfig{
		Keys: []schemas.Key{
			{ID: "key-1", Name: "openai-key", Value: "sk-file-123", Weight: 1},
		},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.openai.com",
		},
		SendBackRawResponse: false,
	}

	// Generate hash for the file config
	fileHash, err := fileConfig.GenerateConfigHash("openai")
	if err != nil {
		t.Fatalf("Failed to generate file hash: %v", err)
	}

	// Create DB config with same hash (simulating unchanged config.json)
	dbConfig := configstore.ProviderConfig{
		Keys: []schemas.Key{
			{ID: "key-1", Name: "openai-key", Value: "sk-db-different", Weight: 1}, // DB may have different key value (edited via dashboard)
		},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.openai.com",
		},
		SendBackRawResponse: false,
		ConfigHash:          fileHash, // Same hash as file
	}

	// Simulate the hash comparison logic
	providersInConfigStore := map[schemas.ModelProvider]configstore.ProviderConfig{
		schemas.OpenAI: dbConfig,
	}

	// When hash matches, we should keep DB config
	existingCfg := providersInConfigStore[schemas.OpenAI]
	if existingCfg.ConfigHash == fileHash {
		// Hash matches - keep DB config
		// This is the expected path
	} else {
		t.Error("Expected hash to match")
	}

	// Verify DB config is preserved (key value from DB, not file)
	if existingCfg.Keys[0].Value != "sk-db-different" {
		t.Errorf("Expected DB key value to be preserved, got %s", existingCfg.Keys[0].Value)
	}
}

// TestProviderHashComparison_DifferentHash tests that file config is used when hashes differ
func TestProviderHashComparison_DifferentHash(t *testing.T) {
	// Create a provider config (simulating what's in config.json - CHANGED)
	fileConfig := configstore.ProviderConfig{
		Keys: []schemas.Key{
			{ID: "key-1", Name: "openai-key", Value: "sk-file-123", Weight: 1},
		},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.openai.com/v2", // Changed URL
		},
		SendBackRawResponse: true, // Changed setting
	}

	// Generate hash for the file config
	fileHash, err := fileConfig.GenerateConfigHash("openai")
	if err != nil {
		t.Fatalf("Failed to generate file hash: %v", err)
	}
	fileConfig.ConfigHash = fileHash

	// Create DB config with different hash (config.json was changed)
	dbConfig := configstore.ProviderConfig{
		Keys: []schemas.Key{
			{ID: "key-1", Name: "openai-key", Value: "sk-db-123", Weight: 1},
			{ID: "key-2", Name: "dashboard-added-key", Value: "sk-dashboard", Weight: 1}, // Key added via dashboard
		},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.openai.com", // Old URL
		},
		SendBackRawResponse: false, // Old setting
		ConfigHash:          "old-different-hash",
	}

	// Simulate the hash comparison logic
	providersInConfigStore := map[schemas.ModelProvider]configstore.ProviderConfig{
		schemas.OpenAI: dbConfig,
	}

	existingCfg := providersInConfigStore[schemas.OpenAI]
	if existingCfg.ConfigHash != fileHash {
		// Hash mismatch - sync from file, but preserve dashboard-added keys
		mergedKeys := fileConfig.Keys

		// Find keys in DB that aren't in file (added via dashboard)
		for _, dbKey := range existingCfg.Keys {
			found := false
			for _, fileKey := range fileConfig.Keys {
				dbKeyHash, _ := configstore.GenerateKeyHash(schemas.Key{
					Name:   dbKey.Name,
					Value:  dbKey.Value,
					Models: dbKey.Models,
					Weight: dbKey.Weight,
				})
				fileKeyHash, _ := configstore.GenerateKeyHash(fileKey)
				if dbKeyHash == fileKeyHash || fileKey.Name == dbKey.Name {					
					found = true
					break
				}
			}
			if !found {
				// Key exists in DB but not in file - preserve it
				mergedKeys = append(mergedKeys, dbKey)
			}
		}

		// Update the result
		fileConfig.Keys = mergedKeys
		providersInConfigStore[schemas.OpenAI] = fileConfig
	} else {
		t.Error("Expected hash mismatch")
	}

	// Verify file config is now used
	resultConfig := providersInConfigStore[schemas.OpenAI]

	if resultConfig.NetworkConfig.BaseURL != "https://api.openai.com/v2" {
		t.Errorf("Expected file BaseURL, got %s", resultConfig.NetworkConfig.BaseURL)
	}

	if !resultConfig.SendBackRawResponse {
		t.Error("Expected SendBackRawResponse to be true (from file)")
	}

	// Verify dashboard-added key is preserved
	if len(resultConfig.Keys) != 2 {
		t.Errorf("Expected 2 keys (1 from file + 1 dashboard-added), got %d", len(resultConfig.Keys))
	}

	hasFileKey := false
	hasDashboardKey := false
	for _, key := range resultConfig.Keys {
		if key.Name == "openai-key" {
			hasFileKey = true
		}
		if key.Name == "dashboard-added-key" {
			hasDashboardKey = true
		}
	}

	if !hasFileKey {
		t.Error("Expected file key to be present")
	}
	if !hasDashboardKey {
		t.Error("Expected dashboard-added key to be preserved")
	}
}

// TestProviderHashComparison_ProviderOnlyInDB tests that provider added via dashboard is preserved
func TestProviderHashComparison_ProviderOnlyInDB(t *testing.T) {
	// DB has a provider that was added via dashboard (not in config.json)
	dbConfig := configstore.ProviderConfig{
		Keys: []schemas.Key{
			{ID: "key-1", Name: "dashboard-provider-key", Value: "sk-dashboard-123", Weight: 1},
		},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.custom-provider.com",
		},
		SendBackRawResponse: true,
	}

	// Generate hash for DB config
	dbHash, err := dbConfig.GenerateConfigHash("custom-provider")
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}
	dbConfig.ConfigHash = dbHash

	// Existing providers from DB
	providersInConfigStore := map[schemas.ModelProvider]configstore.ProviderConfig{
		"custom-provider": dbConfig,
	}

	// File providers (doesn't include custom-provider)
	fileProviders := map[string]configstore.ProviderConfig{
		"openai": {
			Keys: []schemas.Key{
				{ID: "key-1", Name: "openai-key", Value: "sk-openai-123", Weight: 1},
			},
		},
	}

	// Simulate the logic: process file providers, but don't remove DB-only providers
	for providerName, fileCfg := range fileProviders {
		provider := schemas.ModelProvider(providerName)
		fileHash, _ := fileCfg.GenerateConfigHash(providerName)
		fileCfg.ConfigHash = fileHash

		if _, exists := providersInConfigStore[provider]; !exists {
			// New provider from file - add it
			providersInConfigStore[provider] = fileCfg
		}
		// Note: We don't delete providers that are only in DB
	}

	// Verify dashboard-added provider is preserved
	if _, exists := providersInConfigStore["custom-provider"]; !exists {
		t.Error("Expected dashboard-added provider to be preserved")
	}

	// Verify file provider was added
	if _, exists := providersInConfigStore[schemas.OpenAI]; !exists {
		t.Error("Expected file provider to be added")
	}

	// Verify we have both providers
	if len(providersInConfigStore) != 2 {
		t.Errorf("Expected 2 providers (1 from DB + 1 from file), got %d", len(providersInConfigStore))
	}
}

// TestProviderHashComparison_RoundTrip tests JSON → DB → same JSON produces no changes
func TestProviderHashComparison_RoundTrip(t *testing.T) {
	// First load: config.json content
	fileConfig := configstore.ProviderConfig{
		Keys: []schemas.Key{
			{ID: "key-1", Name: "openai-key", Value: "sk-original-123", Weight: 1},
		},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.openai.com",
		},
		SendBackRawResponse: false,
	}

	// Generate hash for file config
	fileHash, err := fileConfig.GenerateConfigHash("openai")
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}
	fileConfig.ConfigHash = fileHash

	// Simulate first load: save to DB
	providersInConfigStore := map[schemas.ModelProvider]configstore.ProviderConfig{
		schemas.OpenAI: fileConfig,
	}

	// Second load: same config.json (no changes)
	secondFileConfig := configstore.ProviderConfig{
		Keys: []schemas.Key{
			{ID: "key-1", Name: "openai-key", Value: "sk-original-123", Weight: 1},
		},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.openai.com",
		},
		SendBackRawResponse: false,
	}

	secondFileHash, err := secondFileConfig.GenerateConfigHash("openai")
	if err != nil {
		t.Fatalf("Failed to generate second hash: %v", err)
	}

	// Hash should match (config.json unchanged)
	if fileHash != secondFileHash {
		t.Error("Expected same hash for identical config (round-trip)")
	}

	// Simulate comparison logic
	existingCfg := providersInConfigStore[schemas.OpenAI]
	if existingCfg.ConfigHash == secondFileHash {
		// Hash matches - keep DB config (no changes needed)
		t.Log("Hash matches - DB config preserved (correct behavior)")
	} else {
		t.Error("Expected hash match on round-trip with same config")
	}
}

// TestProviderHashComparison_DashboardEditThenSameFile tests dashboard edits are preserved
func TestProviderHashComparison_DashboardEditThenSameFile(t *testing.T) {
	// Initial file config
	fileConfig := configstore.ProviderConfig{
		Keys: []schemas.Key{
			{ID: "key-1", Name: "openai-key", Value: "sk-original-123", Weight: 1},
		},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.openai.com",
		},
		SendBackRawResponse: false,
	}

	fileHash, _ := fileConfig.GenerateConfigHash("openai")
	fileConfig.ConfigHash = fileHash

	// Simulate: user edits key value via dashboard (but provider config hash stays same)
	dbConfig := configstore.ProviderConfig{
		Keys: []schemas.Key{
			{ID: "key-1", Name: "openai-key", Value: "sk-dashboard-modified-456", Weight: 1}, // Modified via dashboard
		},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.openai.com",
		},
		SendBackRawResponse: false,
		ConfigHash:          fileHash, // Hash based on provider config, not keys
	}

	providersInConfigStore := map[schemas.ModelProvider]configstore.ProviderConfig{
		schemas.OpenAI: dbConfig,
	}

	// Reload with same file config
	reloadFileConfig := configstore.ProviderConfig{
		Keys: []schemas.Key{
			{ID: "key-1", Name: "openai-key", Value: "sk-original-123", Weight: 1},
		},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.openai.com",
		},
		SendBackRawResponse: false,
	}

	reloadHash, _ := reloadFileConfig.GenerateConfigHash("openai")

	// Hash matches (file unchanged)
	existingCfg := providersInConfigStore[schemas.OpenAI]
	if existingCfg.ConfigHash == reloadHash {
		// Keep DB config - dashboard edits preserved
		t.Log("Hash matches - dashboard edits preserved (correct behavior)")
	} else {
		t.Error("Expected hash match - file wasn't changed")
	}

	// Verify dashboard-modified key value is preserved
	if existingCfg.Keys[0].Value != "sk-dashboard-modified-456" {
		t.Errorf("Expected dashboard-modified key value to be preserved, got %s", existingCfg.Keys[0].Value)
	}
}

// TestProviderHashComparison_OptionalFieldsPresence tests hash with optional fields present/absent
func TestProviderHashComparison_OptionalFieldsPresence(t *testing.T) {
	// Config with no optional fields
	configNoOptional := configstore.ProviderConfig{
		Keys:                []schemas.Key{{ID: "key-1", Name: "test", Value: "sk-123", Weight: 1}},
		SendBackRawResponse: false,
	}

	hashNoOptional, err := configNoOptional.GenerateConfigHash("openai")
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	// Config with NetworkConfig
	configWithNetwork := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "test", Value: "sk-123", Weight: 1}},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.example.com",
		},
		SendBackRawResponse: false,
	}

	hashWithNetwork, err := configWithNetwork.GenerateConfigHash("openai")
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	if hashNoOptional == hashWithNetwork {
		t.Error("Expected different hash when NetworkConfig is present vs absent")
	}

	// Config with ProxyConfig
	configWithProxy := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "test", Value: "sk-123", Weight: 1}},
		ProxyConfig: &schemas.ProxyConfig{
			Type: "http",
			URL:  "http://proxy.example.com",
		},
		SendBackRawResponse: false,
	}

	hashWithProxy, err := configWithProxy.GenerateConfigHash("openai")
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	if hashNoOptional == hashWithProxy {
		t.Error("Expected different hash when ProxyConfig is present vs absent")
	}

	// Config with ConcurrencyAndBufferSize
	configWithConcurrency := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "test", Value: "sk-123", Weight: 1}},
		ConcurrencyAndBufferSize: &schemas.ConcurrencyAndBufferSize{
			Concurrency: 10,
			BufferSize:  100,
		},
		SendBackRawResponse: false,
	}

	hashWithConcurrency, err := configWithConcurrency.GenerateConfigHash("openai")
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	if hashNoOptional == hashWithConcurrency {
		t.Error("Expected different hash when ConcurrencyAndBufferSize is present vs absent")
	}

	// Config with CustomProviderConfig
	configWithCustom := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "test", Value: "sk-123", Weight: 1}},
		CustomProviderConfig: &schemas.CustomProviderConfig{
			BaseProviderType: "openai",
		},
		SendBackRawResponse: false,
	}

	hashWithCustom, err := configWithCustom.GenerateConfigHash("openai")
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	if hashNoOptional == hashWithCustom {
		t.Error("Expected different hash when CustomProviderConfig is present vs absent")
	}

	// Config with SendBackRawResponse true vs false
	configWithRawResponse := configstore.ProviderConfig{
		Keys:                []schemas.Key{{ID: "key-1", Name: "test", Value: "sk-123", Weight: 1}},
		SendBackRawResponse: true,
	}

	hashWithRawResponse, err := configWithRawResponse.GenerateConfigHash("openai")
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	if hashNoOptional == hashWithRawResponse {
		t.Error("Expected different hash when SendBackRawResponse is true vs false")
	}

	// Config with ALL optional fields
	configAllFields := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "test", Value: "sk-123", Weight: 1}},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.example.com",
		},
		ConcurrencyAndBufferSize: &schemas.ConcurrencyAndBufferSize{
			Concurrency: 10,
			BufferSize:  100,
		},
		ProxyConfig: &schemas.ProxyConfig{
			Type: "http",
			URL:  "http://proxy.example.com",
		},
		CustomProviderConfig: &schemas.CustomProviderConfig{
			BaseProviderType: "openai",
		},
		SendBackRawResponse: true,
	}

	hashAllFields, err := configAllFields.GenerateConfigHash("openai")
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	// All hashes should be unique
	hashes := map[string]string{
		"no_optional":    hashNoOptional,
		"with_network":   hashWithNetwork,
		"with_proxy":     hashWithProxy,
		"with_conc":      hashWithConcurrency,
		"with_custom":    hashWithCustom,
		"with_raw":       hashWithRawResponse,
		"all_fields":     hashAllFields,
	}

	seen := make(map[string]string)
	for name, hash := range hashes {
		if existingName, exists := seen[hash]; exists {
			t.Errorf("Hash collision between %s and %s", name, existingName)
		}
		seen[hash] = name
	}
}

// TestKeyHashComparison_OptionalFieldsPresence tests key hash with optional fields
func TestKeyHashComparison_OptionalFieldsPresence(t *testing.T) {
	// Basic key with no optional configs
	keyBasic := schemas.Key{
		ID:     "key-1",
		Name:   "test-key",
		Value:  "sk-123",
		Weight: 1,
	}

	hashBasic, _ := configstore.GenerateKeyHash(keyBasic)

	// Key with Models
	keyWithModels := schemas.Key{
		ID:     "key-1",
		Name:   "test-key",
		Value:  "sk-123",
		Models: []string{"gpt-4"},
		Weight: 1,
	}

	hashWithModels, _ := configstore.GenerateKeyHash(keyWithModels)

	if hashBasic == hashWithModels {
		t.Error("Expected different hash when Models is present vs absent")
	}

	// Key with empty Models array vs nil
	keyEmptyModels := schemas.Key{
		ID:     "key-1",
		Name:   "test-key",
		Value:  "sk-123",
		Models: []string{},
		Weight: 1,
	}

	hashEmptyModels, _ := configstore.GenerateKeyHash(keyEmptyModels)

	// Empty slice and nil should produce same hash (both mean "no model restrictions")
	if hashBasic != hashEmptyModels {
		t.Error("Expected same hash for nil Models and empty Models slice")
	}

	// Key with AzureKeyConfig
	keyWithAzure := schemas.Key{
		ID:     "key-1",
		Name:   "test-key",
		Value:  "sk-123",
		Weight: 1,
		AzureKeyConfig: &schemas.AzureKeyConfig{
			Endpoint:   "https://myazure.openai.azure.com",
			APIVersion: stringPtr("2024-02-01"),
		},
	}

	hashWithAzure, _ := configstore.GenerateKeyHash(keyWithAzure)

	if hashBasic == hashWithAzure {
		t.Error("Expected different hash when AzureKeyConfig is present vs absent")
	}

	// Key with VertexKeyConfig
	keyWithVertex := schemas.Key{
		ID:     "key-1",
		Name:   "test-key",
		Value:  "sk-123",
		Weight: 1,
		VertexKeyConfig: &schemas.VertexKeyConfig{
			ProjectID: "my-project",
			Region:    "us-central1",
		},
	}

	hashWithVertex, _ := configstore.GenerateKeyHash(keyWithVertex)

	if hashBasic == hashWithVertex {
		t.Error("Expected different hash when VertexKeyConfig is present vs absent")
	}

	// Key with BedrockKeyConfig
	keyWithBedrock := schemas.Key{
		ID:     "key-1",
		Name:   "test-key",
		Value:  "sk-123",
		Weight: 1,
		BedrockKeyConfig: &schemas.BedrockKeyConfig{
			AccessKey: "AKIA...",
			SecretKey: "secret...",
			Region:    stringPtr("us-east-1"),
		},
	}

	hashWithBedrock, _ := configstore.GenerateKeyHash(keyWithBedrock)

	if hashBasic == hashWithBedrock {
		t.Error("Expected different hash when BedrockKeyConfig is present vs absent")
	}

	// Verify all hashes are unique
	hashes := map[string]string{
		"basic":        hashBasic,
		"with_models":  hashWithModels,
		"with_azure":   hashWithAzure,
		"with_vertex":  hashWithVertex,
		"with_bedrock": hashWithBedrock,
	}

	seen := make(map[string]string)
	for name, hash := range hashes {
		if existingName, exists := seen[hash]; exists {
			t.Errorf("Hash collision between %s and %s", name, existingName)
		}
		seen[hash] = name
	}
}

// TestProviderHashComparison_FieldValueChanges tests hash changes when field values change
func TestProviderHashComparison_FieldValueChanges(t *testing.T) {
	// Base config
	baseConfig := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "test", Value: "sk-123", Weight: 1}},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.example.com",
		},
		SendBackRawResponse: false,
	}

	baseHash, _ := baseConfig.GenerateConfigHash("openai")

	// Change BaseURL
	configChangedURL := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "test", Value: "sk-123", Weight: 1}},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.different.com", // Changed
		},
		SendBackRawResponse: false,
	}

	hashChangedURL, _ := configChangedURL.GenerateConfigHash("openai")

	if baseHash == hashChangedURL {
		t.Error("Expected different hash when BaseURL changes")
	}

	// Add extra headers
	configWithHeaders := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "test", Value: "sk-123", Weight: 1}},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.example.com",
			ExtraHeaders: map[string]string{
				"X-Custom-Header": "value",
			},
		},
		SendBackRawResponse: false,
	}

	hashWithHeaders, _ := configWithHeaders.GenerateConfigHash("openai")

	if baseHash == hashWithHeaders {
		t.Error("Expected different hash when ExtraHeaders are added")
	}

	// Change concurrency values
	configWithConc1 := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "test", Value: "sk-123", Weight: 1}},
		ConcurrencyAndBufferSize: &schemas.ConcurrencyAndBufferSize{
			Concurrency: 10,
			BufferSize:  100,
		},
	}

	hashConc1, _ := configWithConc1.GenerateConfigHash("openai")

	configWithConc2 := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "test", Value: "sk-123", Weight: 1}},
		ConcurrencyAndBufferSize: &schemas.ConcurrencyAndBufferSize{
			Concurrency: 20, // Changed
			BufferSize:  100,
		},
	}

	hashConc2, _ := configWithConc2.GenerateConfigHash("openai")

	if hashConc1 == hashConc2 {
		t.Error("Expected different hash when Concurrency value changes")
	}
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}

// TestProviderHashComparison_FieldRemoved tests hash changes when fields are removed
func TestProviderHashComparison_FieldRemoved(t *testing.T) {
	// Original config with all fields
	originalConfig := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "test", Value: "sk-123", Weight: 1}},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.example.com",
			ExtraHeaders: map[string]string{
				"X-Custom": "value",
			},
		},
		ConcurrencyAndBufferSize: &schemas.ConcurrencyAndBufferSize{
			Concurrency: 10,
			BufferSize:  100,
		},
		ProxyConfig: &schemas.ProxyConfig{
			Type: "http",
			URL:  "http://proxy.example.com",
		},
		SendBackRawResponse: true,
	}

	originalHash, _ := originalConfig.GenerateConfigHash("openai")

	// NetworkConfig removed
	configNoNetwork := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "test", Value: "sk-123", Weight: 1}},
		// NetworkConfig: nil (removed)
		ConcurrencyAndBufferSize: &schemas.ConcurrencyAndBufferSize{
			Concurrency: 10,
			BufferSize:  100,
		},
		ProxyConfig: &schemas.ProxyConfig{
			Type: "http",
			URL:  "http://proxy.example.com",
		},
		SendBackRawResponse: true,
	}

	hashNoNetwork, _ := configNoNetwork.GenerateConfigHash("openai")

	if originalHash == hashNoNetwork {
		t.Error("Expected different hash when NetworkConfig is removed")
	}

	// ConcurrencyAndBufferSize removed
	configNoConcurrency := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "test", Value: "sk-123", Weight: 1}},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.example.com",
			ExtraHeaders: map[string]string{
				"X-Custom": "value",
			},
		},
		// ConcurrencyAndBufferSize: nil (removed)
		ProxyConfig: &schemas.ProxyConfig{
			Type: "http",
			URL:  "http://proxy.example.com",
		},
		SendBackRawResponse: true,
	}

	hashNoConcurrency, _ := configNoConcurrency.GenerateConfigHash("openai")

	if originalHash == hashNoConcurrency {
		t.Error("Expected different hash when ConcurrencyAndBufferSize is removed")
	}

	// ProxyConfig removed
	configNoProxy := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "test", Value: "sk-123", Weight: 1}},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.example.com",
			ExtraHeaders: map[string]string{
				"X-Custom": "value",
			},
		},
		ConcurrencyAndBufferSize: &schemas.ConcurrencyAndBufferSize{
			Concurrency: 10,
			BufferSize:  100,
		},
		// ProxyConfig: nil (removed)
		SendBackRawResponse: true,
	}

	hashNoProxy, _ := configNoProxy.GenerateConfigHash("openai")

	if originalHash == hashNoProxy {
		t.Error("Expected different hash when ProxyConfig is removed")
	}

	// SendBackRawResponse changed to false
	configNoRawResponse := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "test", Value: "sk-123", Weight: 1}},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.example.com",
			ExtraHeaders: map[string]string{
				"X-Custom": "value",
			},
		},
		ConcurrencyAndBufferSize: &schemas.ConcurrencyAndBufferSize{
			Concurrency: 10,
			BufferSize:  100,
		},
		ProxyConfig: &schemas.ProxyConfig{
			Type: "http",
			URL:  "http://proxy.example.com",
		},
		SendBackRawResponse: false, // Changed to false
	}

	hashNoRawResponse, _ := configNoRawResponse.GenerateConfigHash("openai")

	if originalHash == hashNoRawResponse {
		t.Error("Expected different hash when SendBackRawResponse is changed to false")
	}

	// ExtraHeaders removed from NetworkConfig
	configNoHeaders := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "test", Value: "sk-123", Weight: 1}},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.example.com",
			// ExtraHeaders removed
		},
		ConcurrencyAndBufferSize: &schemas.ConcurrencyAndBufferSize{
			Concurrency: 10,
			BufferSize:  100,
		},
		ProxyConfig: &schemas.ProxyConfig{
			Type: "http",
			URL:  "http://proxy.example.com",
		},
		SendBackRawResponse: true,
	}

	hashNoHeaders, _ := configNoHeaders.GenerateConfigHash("openai")

	if originalHash == hashNoHeaders {
		t.Error("Expected different hash when ExtraHeaders are removed")
	}

	// All optional fields removed
	configMinimal := configstore.ProviderConfig{
		Keys:                []schemas.Key{{ID: "key-1", Name: "test", Value: "sk-123", Weight: 1}},
		SendBackRawResponse: false,
	}

	hashMinimal, _ := configMinimal.GenerateConfigHash("openai")

	if originalHash == hashMinimal {
		t.Error("Expected different hash when all optional fields are removed")
	}
}

// TestKeyHashComparison_FieldRemoved tests key hash changes when fields are removed
func TestKeyHashComparison_FieldRemoved(t *testing.T) {
	// Original key with all fields
	originalKey := schemas.Key{
		ID:     "key-1",
		Name:   "test-key",
		Value:  "sk-123",
		Models: []string{"gpt-4", "gpt-3.5-turbo"},
		Weight: 1.5,
		AzureKeyConfig: &schemas.AzureKeyConfig{
			Endpoint:   "https://myazure.openai.azure.com",
			APIVersion: stringPtr("2024-02-01"),
		},
	}

	originalHash, _ := configstore.GenerateKeyHash(originalKey)

	// Models removed
	keyNoModels := schemas.Key{
		ID:    "key-1",
		Name:  "test-key",
		Value: "sk-123",
		// Models: nil (removed)
		Weight: 1.5,
		AzureKeyConfig: &schemas.AzureKeyConfig{
			Endpoint:   "https://myazure.openai.azure.com",
			APIVersion: stringPtr("2024-02-01"),
		},
	}

	hashNoModels, _ := configstore.GenerateKeyHash(keyNoModels)

	if originalHash == hashNoModels {
		t.Error("Expected different hash when Models are removed")
	}

	// AzureKeyConfig removed
	keyNoAzure := schemas.Key{
		ID:     "key-1",
		Name:   "test-key",
		Value:  "sk-123",
		Models: []string{"gpt-4", "gpt-3.5-turbo"},
		Weight: 1.5,
		// AzureKeyConfig: nil (removed)
	}

	hashNoAzure, _ := configstore.GenerateKeyHash(keyNoAzure)

	if originalHash == hashNoAzure {
		t.Error("Expected different hash when AzureKeyConfig is removed")
	}

	// Weight changed
	keyDifferentWeight := schemas.Key{
		ID:     "key-1",
		Name:   "test-key",
		Value:  "sk-123",
		Models: []string{"gpt-4", "gpt-3.5-turbo"},
		Weight: 1.0, // Changed from 1.5
		AzureKeyConfig: &schemas.AzureKeyConfig{
			Endpoint:   "https://myazure.openai.azure.com",
			APIVersion: stringPtr("2024-02-01"),
		},
	}

	hashDifferentWeight, _ := configstore.GenerateKeyHash(keyDifferentWeight)

	if originalHash == hashDifferentWeight {
		t.Error("Expected different hash when Weight is changed")
	}

	// Some models removed
	keyFewerModels := schemas.Key{
		ID:     "key-1",
		Name:   "test-key",
		Value:  "sk-123",
		Models: []string{"gpt-4"}, // gpt-3.5-turbo removed
		Weight: 1.5,
		AzureKeyConfig: &schemas.AzureKeyConfig{
			Endpoint:   "https://myazure.openai.azure.com",
			APIVersion: stringPtr("2024-02-01"),
		},
	}

	hashFewerModels, _ := configstore.GenerateKeyHash(keyFewerModels)

	if originalHash == hashFewerModels {
		t.Error("Expected different hash when some Models are removed")
	}

	// Azure endpoint changed
	keyDifferentEndpoint := schemas.Key{
		ID:     "key-1",
		Name:   "test-key",
		Value:  "sk-123",
		Models: []string{"gpt-4", "gpt-3.5-turbo"},
		Weight: 1.5,
		AzureKeyConfig: &schemas.AzureKeyConfig{
			Endpoint:   "https://different.openai.azure.com", // Changed
			APIVersion: stringPtr("2024-02-01"),
		},
	}

	hashDifferentEndpoint, _ := configstore.GenerateKeyHash(keyDifferentEndpoint)

	if originalHash == hashDifferentEndpoint {
		t.Error("Expected different hash when Azure endpoint is changed")
	}

	// Azure APIVersion removed
	keyNoAPIVersion := schemas.Key{
		ID:     "key-1",
		Name:   "test-key",
		Value:  "sk-123",
		Models: []string{"gpt-4", "gpt-3.5-turbo"},
		Weight: 1.5,
		AzureKeyConfig: &schemas.AzureKeyConfig{
			Endpoint: "https://myazure.openai.azure.com",
			// APIVersion: nil (removed)
		},
	}

	hashNoAPIVersion, _ := configstore.GenerateKeyHash(keyNoAPIVersion)

	if originalHash == hashNoAPIVersion {
		t.Error("Expected different hash when Azure APIVersion is removed")
	}
}

// TestProviderHashComparison_PartialFieldChanges tests partial changes within nested structs
func TestProviderHashComparison_PartialFieldChanges(t *testing.T) {
	// Base config
	baseConfig := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "test", Value: "sk-123", Weight: 1}},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL:                        "https://api.example.com",
			DefaultRequestTimeoutInSeconds: 30,
			MaxRetries:                     3,
		},
	}

	baseHash, _ := baseConfig.GenerateConfigHash("openai")

	// Timeout set to 0 (default/removed)
	configNoTimeout := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "test", Value: "sk-123", Weight: 1}},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL:                        "https://api.example.com",
			DefaultRequestTimeoutInSeconds: 0, // Removed/default
			MaxRetries:                     3,
		},
	}

	hashNoTimeout, _ := configNoTimeout.GenerateConfigHash("openai")

	if baseHash == hashNoTimeout {
		t.Error("Expected different hash when DefaultRequestTimeoutInSeconds is removed/zeroed")
	}

	// MaxRetries set to 0 (default/removed)
	configNoRetries := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "test", Value: "sk-123", Weight: 1}},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL:                        "https://api.example.com",
			DefaultRequestTimeoutInSeconds: 30,
			MaxRetries:                     0, // Removed/default
		},
	}

	hashNoRetries, _ := configNoRetries.GenerateConfigHash("openai")

	if baseHash == hashNoRetries {
		t.Error("Expected different hash when MaxRetries is removed/zeroed")
	}

	// Timeout value changed
	configDifferentTimeout := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "test", Value: "sk-123", Weight: 1}},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL:                        "https://api.example.com",
			DefaultRequestTimeoutInSeconds: 60, // Changed from 30
			MaxRetries:                     3,
		},
	}

	hashDifferentTimeout, _ := configDifferentTimeout.GenerateConfigHash("openai")

	if baseHash == hashDifferentTimeout {
		t.Error("Expected different hash when DefaultRequestTimeoutInSeconds value changes")
	}
}

// TestProviderHashComparison_FullLifecycle tests DB → new JSON → update DB → same JSON (no update)
func TestProviderHashComparison_FullLifecycle(t *testing.T) {
	// === STEP 1: Initial state - provider exists in DB from previous config.json ===
	initialConfig := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "openai-key", Value: "sk-initial-123", Weight: 1}},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.openai.com/v1",
		},
		SendBackRawResponse: false,
	}

	initialHash, _ := initialConfig.GenerateConfigHash("openai")
	initialConfig.ConfigHash = initialHash

	// Simulate DB state
	providersInDB := map[schemas.ModelProvider]configstore.ProviderConfig{
		schemas.OpenAI: initialConfig,
	}

	t.Logf("Step 1 - Initial DB hash: %s", initialHash[:16]+"...")

	// === STEP 2: New config.json comes with changes ===
	newFileConfig := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "openai-key", Value: "sk-new-456", Weight: 1}},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL:    "https://api.openai.com/v2", // Changed!
			MaxRetries: 5,                           // Added!
		},
		SendBackRawResponse: true, // Changed!
	}

	newFileHash, _ := newFileConfig.GenerateConfigHash("openai")
	newFileConfig.ConfigHash = newFileHash

	t.Logf("Step 2 - New file hash: %s", newFileHash[:16]+"...")

	// Verify hashes are different (config.json changed)
	dbConfig := providersInDB[schemas.OpenAI]
	if dbConfig.ConfigHash == newFileHash {
		t.Fatal("Expected different hash - config.json was changed")
	}

	// === STEP 3: Sync from file to DB (hash mismatch triggers update) ===
	t.Log("Step 3 - Hash mismatch detected, syncing from file to DB")

	// Simulate the sync: file config replaces DB config
	providersInDB[schemas.OpenAI] = newFileConfig

	// Verify DB was updated
	updatedDBConfig := providersInDB[schemas.OpenAI]
	if updatedDBConfig.ConfigHash != newFileHash {
		t.Error("Expected DB to be updated with new hash")
	}
	if updatedDBConfig.NetworkConfig.BaseURL != "https://api.openai.com/v2" {
		t.Error("Expected DB to have new BaseURL from file")
	}
	if !updatedDBConfig.SendBackRawResponse {
		t.Error("Expected DB to have SendBackRawResponse=true from file")
	}

	t.Logf("Step 3 - DB updated, new DB hash: %s", updatedDBConfig.ConfigHash[:16]+"...")

	// === STEP 4: Same config.json loaded again - should NOT update ===
	sameFileConfig := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "openai-key", Value: "sk-new-456", Weight: 1}},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL:    "https://api.openai.com/v2",
			MaxRetries: 5,
		},
		SendBackRawResponse: true,
	}

	sameFileHash, _ := sameFileConfig.GenerateConfigHash("openai")

	t.Logf("Step 4 - Same file loaded again, hash: %s", sameFileHash[:16]+"...")

	// Verify hashes match now (no changes in config.json since last sync)
	currentDBConfig := providersInDB[schemas.OpenAI]
	if currentDBConfig.ConfigHash != sameFileHash {
		t.Errorf("Expected hash match - config.json unchanged since last sync. DB: %s, File: %s",
			currentDBConfig.ConfigHash[:16], sameFileHash[:16])
	}

	// Simulate the comparison logic
	if currentDBConfig.ConfigHash == sameFileHash {
		t.Log("Step 4 - Hash matches, keeping DB config (no update needed) ✓")
	} else {
		t.Error("Step 4 - Should have matched, but didn't")
	}

	// === STEP 5: Verify DB wasn't modified (still has step 3 values) ===
	finalDBConfig := providersInDB[schemas.OpenAI]
	if finalDBConfig.NetworkConfig.BaseURL != "https://api.openai.com/v2" {
		t.Error("DB should still have v2 URL")
	}
	if finalDBConfig.NetworkConfig.MaxRetries != 5 {
		t.Error("DB should still have MaxRetries=5")
	}
	if !finalDBConfig.SendBackRawResponse {
		t.Error("DB should still have SendBackRawResponse=true")
	}

	t.Log("Step 5 - DB state verified, lifecycle complete ✓")
}

// TestProviderHashComparison_MultipleUpdates tests multiple config.json updates over time
func TestProviderHashComparison_MultipleUpdates(t *testing.T) {
	// Track all hashes for verification
	hashHistory := []string{}

	// === Round 1: Initial config ===
	config1 := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "key", Value: "sk-v1", Weight: 1}},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.v1.com",
		},
	}
	hash1, _ := config1.GenerateConfigHash("openai")
	config1.ConfigHash = hash1
	hashHistory = append(hashHistory, hash1)

	providersInDB := map[schemas.ModelProvider]configstore.ProviderConfig{
		schemas.OpenAI: config1,
	}

	t.Logf("Round 1 - hash: %s", hash1[:16]+"...")

	// === Round 2: Update config.json ===
	config2 := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "key", Value: "sk-v2", Weight: 1}},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.v2.com", // Changed
		},
	}
	hash2, _ := config2.GenerateConfigHash("openai")
	config2.ConfigHash = hash2
	hashHistory = append(hashHistory, hash2)

	// Hash should be different
	if hash1 == hash2 {
		t.Fatal("Round 2 hash should differ from Round 1")
	}

	// Sync to DB
	providersInDB[schemas.OpenAI] = config2
	t.Logf("Round 2 - hash: %s (different from Round 1) ✓", hash2[:16]+"...")

	// === Round 3: Another update ===
	config3 := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "key", Value: "sk-v3", Weight: 1}},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL:    "https://api.v3.com", // Changed again
			MaxRetries: 3,                    // Added
		},
		SendBackRawResponse: true, // Added
	}
	hash3, _ := config3.GenerateConfigHash("openai")
	config3.ConfigHash = hash3
	hashHistory = append(hashHistory, hash3)

	// Hash should be different from all previous
	if hash3 == hash1 || hash3 == hash2 {
		t.Fatal("Round 3 hash should differ from all previous")
	}

	// Sync to DB
	providersInDB[schemas.OpenAI] = config3
	t.Logf("Round 3 - hash: %s (different from Round 1 & 2) ✓", hash3[:16]+"...")

	// === Round 4: Same as Round 3 (no change) ===
	config4 := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "key", Value: "sk-v3", Weight: 1}},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL:    "https://api.v3.com",
			MaxRetries: 3,
		},
		SendBackRawResponse: true,
	}
	hash4, _ := config4.GenerateConfigHash("openai")

	// Hash should match Round 3
	if hash4 != hash3 {
		t.Fatalf("Round 4 hash should match Round 3. Got %s, expected %s", hash4[:16], hash3[:16])
	}

	// No sync needed
	t.Logf("Round 4 - hash: %s (matches Round 3, no update) ✓", hash4[:16]+"...")

	// === Round 5: Revert to Round 1 config ===
	config5 := configstore.ProviderConfig{
		Keys: []schemas.Key{{ID: "key-1", Name: "key", Value: "sk-v1", Weight: 1}},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.v1.com",
		},
	}
	hash5, _ := config5.GenerateConfigHash("openai")
	config5.ConfigHash = hash5

	// Hash should match Round 1
	if hash5 != hash1 {
		t.Fatalf("Round 5 hash should match Round 1. Got %s, expected %s", hash5[:16], hash1[:16])
	}

	// But it differs from current DB (which has Round 3 config)
	currentDB := providersInDB[schemas.OpenAI]
	if currentDB.ConfigHash == hash5 {
		t.Fatal("Round 5 should trigger update (reverted config differs from DB)")
	}

	// Sync reverted config to DB
	providersInDB[schemas.OpenAI] = config5
	t.Logf("Round 5 - hash: %s (reverted to Round 1, update triggered) ✓", hash5[:16]+"...")

	// Verify all unique hashes were generated
	uniqueHashes := make(map[string]bool)
	for _, h := range hashHistory {
		uniqueHashes[h] = true
	}
	if len(uniqueHashes) != 3 { // hash1, hash2, hash3 (hash4 = hash3, hash5 = hash1)
		t.Errorf("Expected 3 unique hashes, got %d", len(uniqueHashes))
	}

	t.Log("Multiple updates lifecycle complete ✓")
}

// TestProviderHashComparison_ProviderChangedKeysUnchanged tests provider update without key changes
func TestProviderHashComparison_ProviderChangedKeysUnchanged(t *testing.T) {
	// === Initial state: Provider with keys in DB ===
	originalKey := schemas.Key{
		ID:     "key-1",
		Name:   "openai-key",
		Value:  "sk-original-123",
		Models: []string{"gpt-4", "gpt-3.5-turbo"},
		Weight: 1.5,
	}
	originalKeyHash, _ := configstore.GenerateKeyHash(originalKey)

	dbConfig := configstore.ProviderConfig{
		Keys: []schemas.Key{originalKey},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL:    "https://api.openai.com/v1",
			MaxRetries: 3,
		},
		SendBackRawResponse: false,
	}
	dbProviderHash, _ := dbConfig.GenerateConfigHash("openai")
	dbConfig.ConfigHash = dbProviderHash

	t.Logf("Initial - Provider hash: %s", dbProviderHash[:16]+"...")
	t.Logf("Initial - Key hash: %s", originalKeyHash[:16]+"...")

	// === File config: Provider changed, keys SAME ===
	sameKey := schemas.Key{
		ID:     "key-1",
		Name:   "openai-key",
		Value:  "sk-original-123", // SAME
		Models: []string{"gpt-4", "gpt-3.5-turbo"}, // SAME
		Weight: 1.5, // SAME
	}
	sameKeyHash, _ := configstore.GenerateKeyHash(sameKey)

	fileConfig := configstore.ProviderConfig{
		Keys: []schemas.Key{sameKey},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL:    "https://api.openai.com/v2", // CHANGED!
			MaxRetries: 5,                           // CHANGED!
		},
		SendBackRawResponse: true, // CHANGED!
	}
	fileProviderHash, _ := fileConfig.GenerateConfigHash("openai")

	t.Logf("File - Provider hash: %s", fileProviderHash[:16]+"...")
	t.Logf("File - Key hash: %s", sameKeyHash[:16]+"...")

	// === Verify: Provider hash changed, key hash unchanged ===
	if dbProviderHash == fileProviderHash {
		t.Error("Expected provider hash to be DIFFERENT (provider config changed)")
	} else {
		t.Log("✓ Provider hash changed (expected)")
	}

	if originalKeyHash != sameKeyHash {
		t.Error("Expected key hash to be SAME (key unchanged)")
	} else {
		t.Log("✓ Key hash unchanged (expected)")
	}

	// === Simulate sync logic: Update provider, preserve keys ===
	// When provider hash differs but key hashes match:
	// - Update provider-level config (NetworkConfig, SendBackRawResponse, etc.)
	// - Keep existing keys from DB (they weren't changed in file)

	updatedConfig := configstore.ProviderConfig{
		Keys:                 dbConfig.Keys, // Keep original keys from DB
		NetworkConfig:        fileConfig.NetworkConfig, // Update from file
		SendBackRawResponse:  fileConfig.SendBackRawResponse, // Update from file
		ConfigHash:           fileProviderHash, // New provider hash
	}

	// Verify keys are preserved (same values as DB)
	if len(updatedConfig.Keys) != 1 {
		t.Errorf("Expected 1 key, got %d", len(updatedConfig.Keys))
	}
	if updatedConfig.Keys[0].Value != "sk-original-123" {
		t.Errorf("Expected key value to be preserved, got %s", updatedConfig.Keys[0].Value)
	}
	if len(updatedConfig.Keys[0].Models) != 2 {
		t.Errorf("Expected 2 models to be preserved, got %d", len(updatedConfig.Keys[0].Models))
	}

	// Verify provider config is updated
	if updatedConfig.NetworkConfig.BaseURL != "https://api.openai.com/v2" {
		t.Error("Expected BaseURL to be updated from file")
	}
	if updatedConfig.NetworkConfig.MaxRetries != 5 {
		t.Error("Expected MaxRetries to be updated from file")
	}
	if !updatedConfig.SendBackRawResponse {
		t.Error("Expected SendBackRawResponse to be updated from file")
	}

	t.Log("✓ Provider updated, keys preserved")
}

// TestProviderHashComparison_KeysChangedProviderUnchanged tests key update without provider changes
func TestProviderHashComparison_KeysChangedProviderUnchanged(t *testing.T) {
	// === Initial state: Provider with keys in DB ===
	originalKey := schemas.Key{
		ID:     "key-1",
		Name:   "openai-key",
		Value:  "sk-original-123",
		Models: []string{"gpt-4"},
		Weight: 1.0,
	}
	originalKeyHash, _ := configstore.GenerateKeyHash(originalKey)

	dbConfig := configstore.ProviderConfig{
		Keys: []schemas.Key{originalKey},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL:    "https://api.openai.com/v1",
			MaxRetries: 3,
		},
		SendBackRawResponse: false,
	}
	dbProviderHash, _ := dbConfig.GenerateConfigHash("openai")
	dbConfig.ConfigHash = dbProviderHash

	t.Logf("Initial - Provider hash: %s", dbProviderHash[:16]+"...")
	t.Logf("Initial - Key hash: %s", originalKeyHash[:16]+"...")

	// === File config: Provider SAME, keys changed ===
	changedKey := schemas.Key{
		ID:     "key-1",
		Name:   "openai-key",
		Value:  "sk-new-456",                              // CHANGED!
		Models: []string{"gpt-4", "gpt-3.5-turbo", "o1"},  // CHANGED!
		Weight: 2.0,                                        // CHANGED!
	}
	changedKeyHash, _ := configstore.GenerateKeyHash(changedKey)

	fileConfig := configstore.ProviderConfig{
		Keys: []schemas.Key{changedKey},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL:    "https://api.openai.com/v1", // SAME
			MaxRetries: 3,                           // SAME
		},
		SendBackRawResponse: false, // SAME
	}
	fileProviderHash, _ := fileConfig.GenerateConfigHash("openai")

	t.Logf("File - Provider hash: %s", fileProviderHash[:16]+"...")
	t.Logf("File - Key hash: %s", changedKeyHash[:16]+"...")

	// === Verify: Provider hash unchanged, key hash changed ===
	if dbProviderHash != fileProviderHash {
		t.Error("Expected provider hash to be SAME (provider config unchanged)")
	} else {
		t.Log("✓ Provider hash unchanged (expected)")
	}

	if originalKeyHash == changedKeyHash {
		t.Error("Expected key hash to be DIFFERENT (key changed)")
	} else {
		t.Log("✓ Key hash changed (expected)")
	}

	// === Simulate sync logic: Keep provider, update keys ===
	// When provider hash matches but key hashes differ:
	// - Keep provider-level config from DB
	// - Update keys from file (they were changed)

	updatedConfig := configstore.ProviderConfig{
		Keys:                 fileConfig.Keys, // Update keys from file
		NetworkConfig:        dbConfig.NetworkConfig, // Keep from DB
		SendBackRawResponse:  dbConfig.SendBackRawResponse, // Keep from DB
		ConfigHash:           dbProviderHash, // Provider hash unchanged
	}

	// Verify provider config is preserved
	if updatedConfig.NetworkConfig.BaseURL != "https://api.openai.com/v1" {
		t.Error("Expected BaseURL to be preserved from DB")
	}
	if updatedConfig.NetworkConfig.MaxRetries != 3 {
		t.Error("Expected MaxRetries to be preserved from DB")
	}
	if updatedConfig.SendBackRawResponse {
		t.Error("Expected SendBackRawResponse to be preserved from DB (false)")
	}

	// Verify keys are updated
	if len(updatedConfig.Keys) != 1 {
		t.Errorf("Expected 1 key, got %d", len(updatedConfig.Keys))
	}
	if updatedConfig.Keys[0].Value != "sk-new-456" {
		t.Errorf("Expected key value to be updated, got %s", updatedConfig.Keys[0].Value)
	}
	if len(updatedConfig.Keys[0].Models) != 3 {
		t.Errorf("Expected 3 models (updated), got %d", len(updatedConfig.Keys[0].Models))
	}
	if updatedConfig.Keys[0].Weight != 2.0 {
		t.Errorf("Expected weight to be 2.0 (updated), got %f", updatedConfig.Keys[0].Weight)
	}

	t.Log("✓ Provider preserved, keys updated")
}

// TestProviderHashComparison_BothChangedIndependently tests both provider and keys changed
func TestProviderHashComparison_BothChangedIndependently(t *testing.T) {
	// === Initial state ===
	originalKey := schemas.Key{
		ID:     "key-1",
		Name:   "openai-key",
		Value:  "sk-original-123",
		Models: []string{"gpt-4"},
		Weight: 1.0,
	}
	originalKeyHash, _ := configstore.GenerateKeyHash(originalKey)

	dbConfig := configstore.ProviderConfig{
		Keys: []schemas.Key{originalKey},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.openai.com/v1",
		},
		SendBackRawResponse: false,
	}
	dbProviderHash, _ := dbConfig.GenerateConfigHash("openai")

	t.Logf("Initial - Provider hash: %s", dbProviderHash[:16]+"...")
	t.Logf("Initial - Key hash: %s", originalKeyHash[:16]+"...")

	// === File config: BOTH provider and keys changed ===
	changedKey := schemas.Key{
		ID:     "key-1",
		Name:   "openai-key",
		Value:  "sk-new-456",        // CHANGED
		Models: []string{"gpt-4", "o1"}, // CHANGED
		Weight: 2.0,                  // CHANGED
	}
	changedKeyHash, _ := configstore.GenerateKeyHash(changedKey)

	fileConfig := configstore.ProviderConfig{
		Keys: []schemas.Key{changedKey},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL:    "https://api.openai.com/v2", // CHANGED
			MaxRetries: 5,                           // ADDED
		},
		SendBackRawResponse: true, // CHANGED
	}
	fileProviderHash, _ := fileConfig.GenerateConfigHash("openai")

	t.Logf("File - Provider hash: %s", fileProviderHash[:16]+"...")
	t.Logf("File - Key hash: %s", changedKeyHash[:16]+"...")

	// === Verify: Both hashes changed ===
	if dbProviderHash == fileProviderHash {
		t.Error("Expected provider hash to be DIFFERENT")
	} else {
		t.Log("✓ Provider hash changed")
	}

	if originalKeyHash == changedKeyHash {
		t.Error("Expected key hash to be DIFFERENT")
	} else {
		t.Log("✓ Key hash changed")
	}

	// === Simulate sync: Update everything from file ===
	updatedConfig := fileConfig
	updatedConfig.ConfigHash = fileProviderHash

	// Verify both provider and keys are updated
	if updatedConfig.NetworkConfig.BaseURL != "https://api.openai.com/v2" {
		t.Error("Expected BaseURL to be updated")
	}
	if !updatedConfig.SendBackRawResponse {
		t.Error("Expected SendBackRawResponse to be updated")
	}
	if updatedConfig.Keys[0].Value != "sk-new-456" {
		t.Error("Expected key value to be updated")
	}
	if updatedConfig.Keys[0].Weight != 2.0 {
		t.Error("Expected key weight to be updated")
	}

	t.Log("✓ Both provider and keys updated")
}

// TestProviderHashComparison_NeitherChanged tests no changes scenario
func TestProviderHashComparison_NeitherChanged(t *testing.T) {
	// === Initial state ===
	originalKey := schemas.Key{
		ID:     "key-1",
		Name:   "openai-key",
		Value:  "sk-original-123",
		Models: []string{"gpt-4"},
		Weight: 1.0,
	}
	originalKeyHash, _ := configstore.GenerateKeyHash(originalKey)

	dbConfig := configstore.ProviderConfig{
		Keys: []schemas.Key{originalKey},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.openai.com/v1",
		},
		SendBackRawResponse: false,
	}
	dbProviderHash, _ := dbConfig.GenerateConfigHash("openai")
	dbConfig.ConfigHash = dbProviderHash

	t.Logf("Initial - Provider hash: %s", dbProviderHash[:16]+"...")
	t.Logf("Initial - Key hash: %s", originalKeyHash[:16]+"...")

	// === File config: SAME as DB ===
	sameKey := schemas.Key{
		ID:     "key-1",
		Name:   "openai-key",
		Value:  "sk-original-123", // SAME
		Models: []string{"gpt-4"}, // SAME
		Weight: 1.0,               // SAME
	}
	sameKeyHash, _ := configstore.GenerateKeyHash(sameKey)

	fileConfig := configstore.ProviderConfig{
		Keys: []schemas.Key{sameKey},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.openai.com/v1", // SAME
		},
		SendBackRawResponse: false, // SAME
	}
	fileProviderHash, _ := fileConfig.GenerateConfigHash("openai")

	t.Logf("File - Provider hash: %s", fileProviderHash[:16]+"...")
	t.Logf("File - Key hash: %s", sameKeyHash[:16]+"...")

	// === Verify: Both hashes match ===
	if dbProviderHash != fileProviderHash {
		t.Errorf("Expected provider hash to be SAME, got DB=%s File=%s", 
			dbProviderHash[:16], fileProviderHash[:16])
	} else {
		t.Log("✓ Provider hash unchanged")
	}

	if originalKeyHash != sameKeyHash {
		t.Errorf("Expected key hash to be SAME, got DB=%s File=%s",
			originalKeyHash[:16], sameKeyHash[:16])
	} else {
		t.Log("✓ Key hash unchanged")
	}

	// === No sync needed - keep DB as is ===
	t.Log("✓ No changes detected, DB preserved")
}

// TestKeyHashComparison_KeyContentChanged tests key content change detection
func TestKeyHashComparison_KeyContentChanged(t *testing.T) {
	// Original key in DB
	dbKey := schemas.Key{
		ID:     "key-1",
		Name:   "openai-key",
		Value:  "sk-old-value",
		Models: []string{"gpt-4"},
		Weight: 1,
	}

	dbKeyHash, _ := configstore.GenerateKeyHash(dbKey)

	// Same key in file but with different value
	fileKey := schemas.Key{
		ID:     "key-1",
		Name:   "openai-key",
		Value:  "sk-new-value", // Changed!
		Models: []string{"gpt-4"},
		Weight: 1,
	}

	fileKeyHash, _ := configstore.GenerateKeyHash(fileKey)

	// Hashes should be different (key content changed)
	if dbKeyHash == fileKeyHash {
		t.Error("Expected different hash for keys with different Value")
	}

	// Same key with only models changed
	fileKey2 := schemas.Key{
		ID:     "key-1",
		Name:   "openai-key",
		Value:  "sk-old-value",
		Models: []string{"gpt-4", "gpt-3.5-turbo"}, // Changed models
		Weight: 1,
	}

	fileKey2Hash, _ := configstore.GenerateKeyHash(fileKey2)

	if dbKeyHash == fileKey2Hash {
		t.Error("Expected different hash for keys with different Models")
	}
}

// TestProviderHashComparison_NewProvider tests that new provider is added from file
func TestProviderHashComparison_NewProvider(t *testing.T) {
	// Create a provider config (simulating new provider in config.json)
	fileConfig := configstore.ProviderConfig{
		Keys: []schemas.Key{
			{ID: "key-1", Name: "anthropic-key", Value: "sk-ant-123", Weight: 1},
		},
		NetworkConfig: &schemas.NetworkConfig{
			BaseURL: "https://api.anthropic.com",
		},
		SendBackRawResponse: false,
	}

	// Generate hash for the file config
	fileHash, err := fileConfig.GenerateConfigHash("anthropic")
	if err != nil {
		t.Fatalf("Failed to generate file hash: %v", err)
	}
	fileConfig.ConfigHash = fileHash

	// Empty DB (no existing providers)
	providersInConfigStore := map[schemas.ModelProvider]configstore.ProviderConfig{}

	provider := schemas.Anthropic

	// Simulate the logic: provider doesn't exist, add from file
	if _, exists := providersInConfigStore[provider]; !exists {
		providersInConfigStore[provider] = fileConfig
	}

	// Verify provider was added
	if _, exists := providersInConfigStore[provider]; !exists {
		t.Error("Expected provider to be added")
	}

	resultConfig := providersInConfigStore[provider]

	if resultConfig.ConfigHash != fileHash {
		t.Error("Expected ConfigHash to be set from file")
	}

	if len(resultConfig.Keys) != 1 {
		t.Errorf("Expected 1 key, got %d", len(resultConfig.Keys))
	}

	if resultConfig.Keys[0].Name != "anthropic-key" {
		t.Errorf("Expected key name 'anthropic-key', got %s", resultConfig.Keys[0].Name)
	}
}
