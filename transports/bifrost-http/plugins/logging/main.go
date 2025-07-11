// Package logging provides a SQLite-based logging plugin for Bifrost.
// This plugin stores comprehensive logs of all requests and responses with search,
// filter, and pagination capabilities.
package logging

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/maximhq/bifrost/core/schemas"
)

const (
	PluginName = "bifrost-http-logging"
)

// ContextKey is a custom type for context keys to prevent collisions
type ContextKey string

// LogOperation represents the type of logging operation
type LogOperation string

const (
	LogOperationCreate LogOperation = "create"
	LogOperationUpdate LogOperation = "update"
)

// Context keys for logging optimization
const (
	DroppedCreateContextKey ContextKey = "bifrost-logging-dropped"
)

// LogMessage represents a message in the logging queue
type LogMessage struct {
	Operation   LogOperation
	RequestID   string
	Timestamp   time.Time       // Of the preHook/postHook call
	InitialData *InitialLogData // For create operations
	UpdateData  *UpdateLogData  // For update operations
}

// InitialLogData contains data for initial log entry creation
type InitialLogData struct {
	Provider     string
	Model        string
	Object       string
	InputHistory []schemas.BifrostMessage
	Params       *schemas.ModelParameters
	Tools        *[]schemas.Tool
}

// UpdateLogData contains data for log entry updates
type UpdateLogData struct {
	Status        string
	TokenUsage    *schemas.LLMUsage
	OutputMessage *schemas.BifrostMessage
	ToolCalls     *[]schemas.ToolCall
	ErrorDetails  *schemas.BifrostError
	Model         string // May be different from request
	Object        string // May be different from request
}

// LogEntry represents a complete log entry for a request/response cycle
type LogEntry struct {
	ID            string                   `json:"id"`
	Timestamp     time.Time                `json:"timestamp"`
	Object        string                   `json:"object"` // text.completion, chat.completion, or embedding
	Provider      string                   `json:"provider"`
	Model         string                   `json:"model"`
	InputHistory  []schemas.BifrostMessage `json:"input_history,omitempty"`
	OutputMessage *schemas.BifrostMessage  `json:"output_message,omitempty"`
	Params        *schemas.ModelParameters `json:"params,omitempty"`
	Tools         *[]schemas.Tool          `json:"tools,omitempty"`
	ToolCalls     *[]schemas.ToolCall      `json:"tool_calls,omitempty"`
	Latency       *float64                 `json:"latency,omitempty"`
	TokenUsage    *schemas.LLMUsage        `json:"token_usage,omitempty"`
	Status        string                   `json:"status"` // "processing", "success", or "error"
	ErrorDetails  *schemas.BifrostError    `json:"error_details,omitempty"`
	CreatedAt     time.Time                `json:"created_at"`
}

// SearchFilters represents the available filters for log searches
type SearchFilters struct {
	Providers     []string   `json:"providers,omitempty"`
	Models        []string   `json:"models,omitempty"`
	Status        []string   `json:"status,omitempty"`
	Objects       []string   `json:"objects,omitempty"` // For filtering by request type (chat.completion, text.completion, embedding)
	StartTime     *time.Time `json:"start_time,omitempty"`
	EndTime       *time.Time `json:"end_time,omitempty"`
	MinLatency    *float64   `json:"min_latency,omitempty"`
	MaxLatency    *float64   `json:"max_latency,omitempty"`
	MinTokens     *int       `json:"min_tokens,omitempty"`
	MaxTokens     *int       `json:"max_tokens,omitempty"`
	ContentSearch string     `json:"content_search,omitempty"`
}

// PaginationOptions represents pagination parameters
type PaginationOptions struct {
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	SortBy string `json:"sort_by"` // "timestamp", "latency", "tokens"
	Order  string `json:"order"`   // "asc", "desc"
}

// SearchResult represents the result of a log search
type SearchResult struct {
	Logs       []LogEntry        `json:"logs"`
	Pagination PaginationOptions `json:"pagination"`
	Stats      struct {
		TotalRequests  int64   `json:"total_requests"`
		SuccessRate    float64 `json:"success_rate"`    // Percentage of successful requests
		AverageLatency float64 `json:"average_latency"` // Average latency in milliseconds
		TotalTokens    int64   `json:"total_tokens"`    // Total tokens used
	} `json:"stats"`
}

// Config represents the configuration for the logging plugin
type Config struct {
	DatabasePath string `json:"database_path"`
	// SQLite memory optimization is now handled via connection string parameters
}

// LogCallback is a function that gets called when a new log entry is created
type LogCallback func(*LogEntry)

// LoggerPlugin implements the schemas.Plugin interface
type LoggerPlugin struct {
	config          *Config
	db              *sql.DB
	mu              sync.Mutex
	done            chan struct{}
	wg              sync.WaitGroup
	logger          schemas.Logger
	logCallback     LogCallback
	droppedRequests atomic.Int64
	cleanupTicker   *time.Ticker // Ticker for cleaning up old processing logs
	logMsgPool      sync.Pool    // Pool for reusing LogMessage structs
}

// NewLoggerPlugin creates a new logging plugin
func NewLoggerPlugin(config *Config, logger schemas.Logger) (*LoggerPlugin, error) {
	if config == nil {
		config = &Config{
			DatabasePath: "./bifrost-logs.db",
		}
	}

	// Handle legacy database path (if it was a directory for BadgerDB)
	dbPath := config.DatabasePath
	if !strings.HasSuffix(dbPath, ".db") {
		dbPath = filepath.Join(dbPath, "logs.db")
	}

	// Ensure the directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory %s: %w", dbDir, err)
	}

	// Open SQLite with optimized settings for low memory usage
	db, err := sql.Open("sqlite3", dbPath+"?cache=shared&_journal_mode=WAL&_synchronous=NORMAL&_auto_vacuum=incremental&_page_size=4096&_temp_store=FILE&_mmap_size=0")
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database at %s: %w", dbPath, err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	plugin := &LoggerPlugin{
		config: config,
		db:     db,
		done:   make(chan struct{}),
		logger: logger,
		logMsgPool: sync.Pool{
			New: func() interface{} {
				return &LogMessage{}
			},
		},
	}

	// Prewarm the log message pool
	for range 1000 {
		plugin.getLogMessage()
	}

	// Create tables and indexes
	if err := plugin.createTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	// Start cleanup ticker (runs every 30 seconds)
	plugin.cleanupTicker = time.NewTicker(30 * time.Second)
	plugin.wg.Add(1)
	go plugin.cleanupWorker()

	return plugin, nil
}

// createTables creates the SQLite tables and indexes
func (p *LoggerPlugin) createTables() error {
	// Main logs table with updated schema
	createTable := `
	CREATE TABLE IF NOT EXISTS logs (
		id TEXT PRIMARY KEY,
		timestamp INTEGER,
		provider TEXT NOT NULL,
		model TEXT NOT NULL,
		object_type TEXT NOT NULL,
		status TEXT NOT NULL,
		latency REAL,
		prompt_tokens INTEGER,
		completion_tokens INTEGER,
		total_tokens INTEGER,
		
		-- Store complex fields as JSON
		input_history TEXT,
		output_message TEXT,
		tools TEXT,
		tool_calls TEXT,
		params TEXT,
		error_details TEXT,
		
		-- For content search
		content_summary TEXT,
		
		-- Timestamps for tracking
		created_at INTEGER NOT NULL
	)`

	if _, err := p.db.Exec(createTable); err != nil {
		return fmt.Errorf("failed to create logs table: %w", err)
	}

	// Check if we need to add the new columns to existing table
	if err := p.migrateTableSchema(); err != nil {
		return fmt.Errorf("failed to migrate table schema: %w", err)
	}

	// Create indexes for fast filtering
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_timestamp ON logs(timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_provider ON logs(provider)",
		"CREATE INDEX IF NOT EXISTS idx_model ON logs(model)",
		"CREATE INDEX IF NOT EXISTS idx_object_type ON logs(object_type)",
		"CREATE INDEX IF NOT EXISTS idx_status ON logs(status)",
		"CREATE INDEX IF NOT EXISTS idx_created_at ON logs(created_at)",
	}

	for _, index := range indexes {
		if _, err := p.db.Exec(index); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Check if FTS5 is available
	var ftsAvailable bool
	err := p.db.QueryRow("SELECT 1 FROM pragma_compile_options WHERE compile_options = 'ENABLE_FTS5'").Scan(&ftsAvailable)
	if err != nil {
		p.logger.Debug("FTS5 not available for logging, falling back to regular search")
	} else {
		createFTS := `
			CREATE VIRTUAL TABLE IF NOT EXISTS logs_fts USING fts5(
				id, content_summary, content='logs', content_rowid='rowid'
			)`

		if _, err := p.db.Exec(createFTS); err != nil {
			p.logger.Warn(fmt.Sprintf("Failed to create FTS table, falling back to LIKE search: %v", err))
		} else {
			// Create triggers to keep FTS table in sync
			triggers := []string{
				`CREATE TRIGGER IF NOT EXISTS logs_fts_insert AFTER INSERT ON logs BEGIN
						INSERT INTO logs_fts(id, content_summary) VALUES (new.id, new.content_summary);
					END`,
				`CREATE TRIGGER IF NOT EXISTS logs_fts_update AFTER UPDATE ON logs BEGIN
						UPDATE logs_fts SET content_summary = new.content_summary WHERE id = new.id;
					END`,
				`CREATE TRIGGER IF NOT EXISTS logs_fts_delete AFTER DELETE ON logs BEGIN
						DELETE FROM logs_fts WHERE id = old.id;
					END`,
			}

			for _, trigger := range triggers {
				if _, err := p.db.Exec(trigger); err != nil {
					p.logger.Warn(fmt.Sprintf("Failed to create FTS trigger: %v", err))
				}
			}
		}
	}

	return nil
}

// migrateTableSchema adds new columns if they don't exist
func (p *LoggerPlugin) migrateTableSchema() error {
	// Check if created_at column exists
	var columnExists bool
	err := p.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('logs') WHERE name = 'created_at'").Scan(&columnExists)
	if err != nil {
		return fmt.Errorf("failed to check for created_at column: %w", err)
	}

	if !columnExists {
		if _, err := p.db.Exec("ALTER TABLE logs ADD COLUMN created_at INTEGER DEFAULT 0"); err != nil {
			return fmt.Errorf("failed to add created_at column: %w", err)
		}
	}

	return nil
}

// cleanupWorker periodically removes old processing logs
func (p *LoggerPlugin) cleanupWorker() {
	defer p.wg.Done()

	for {
		select {
		case <-p.cleanupTicker.C:
			p.cleanupOldProcessingLogs()

		case <-p.done:
			return
		}
	}
}

// cleanupOldProcessingLogs removes processing logs older than 1 minute
func (p *LoggerPlugin) cleanupOldProcessingLogs() {
	// Calculate timestamp for 1 minute ago
	oneMinuteAgo := time.Now().Add(-1 * time.Minute).UnixNano()

	// Delete processing logs older than 1 minute
	query := `DELETE FROM logs WHERE status = 'processing' AND created_at < ?`
	result, err := p.db.Exec(query, oneMinuteAgo)
	if err != nil {
		p.logger.Error(fmt.Errorf("failed to cleanup old processing logs: %w", err))
		return
	}

	// Log the cleanup activity
	if rowsAffected, err := result.RowsAffected(); err == nil && rowsAffected > 0 {
		p.logger.Debug(fmt.Sprintf("Cleaned up %d old processing logs", rowsAffected))
	}
}

// getLogMessage gets a LogMessage from the pool
func (p *LoggerPlugin) getLogMessage() *LogMessage {
	return p.logMsgPool.Get().(*LogMessage)
}

// putLogMessage returns a LogMessage to the pool after resetting it
func (p *LoggerPlugin) putLogMessage(msg *LogMessage) {
	// Reset the message fields to avoid memory leaks
	msg.Operation = ""
	msg.RequestID = ""
	msg.Timestamp = time.Time{}
	msg.InitialData = nil
	msg.UpdateData = nil

	p.logMsgPool.Put(msg)
}

// SetLogCallback sets a callback function that will be called for each log entry
func (p *LoggerPlugin) SetLogCallback(callback LogCallback) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.logCallback = callback
}

// GetName returns the name of the plugin
func (p *LoggerPlugin) GetName() string {
	return PluginName
}

// PreHook is called before a request is processed - FULLY ASYNC, NO DATABASE I/O
func (p *LoggerPlugin) PreHook(ctx *context.Context, req *schemas.BifrostRequest) (*schemas.BifrostRequest, *schemas.PluginShortCircuit, error) {
	if ctx == nil {
		// Log error but don't fail the request
		p.logger.Error(fmt.Errorf("context is nil in PreHook"))
		return req, nil, nil
	}

	// Extract request ID from context
	requestID, ok := (*ctx).Value(ContextKey("request-id")).(string)
	if !ok || requestID == "" {
		// Log error but don't fail the request
		p.logger.Error(fmt.Errorf("request-id not found in context or is empty"))
		return req, nil, nil
	}

	// Prepare initial log data
	objectType := p.determineObjectType(req.Input)
	inputHistory := p.extractInputHistory(req.Input)

	initialData := &InitialLogData{
		Provider:     string(req.Provider),
		Model:        req.Model,
		Object:       objectType,
		InputHistory: inputHistory,
		Params:       req.Params,
	}

	if req.Params != nil && req.Params.Tools != nil {
		initialData.Tools = req.Params.Tools
	}

	// Queue the log creation message (non-blocking) - Using sync.Pool
	logMsg := p.getLogMessage()
	logMsg.Operation = LogOperationCreate
	logMsg.RequestID = requestID
	logMsg.Timestamp = time.Now()
	logMsg.InitialData = initialData

	go func(logMsg *LogMessage) {
		defer p.putLogMessage(logMsg) // Return to pool when done
		if err := p.insertInitialLogEntry(logMsg.RequestID, logMsg.Timestamp, logMsg.InitialData); err != nil {
			p.logger.Error(fmt.Errorf("failed to insert initial log entry for request %s: %w", logMsg.RequestID, err))
		}
	}(logMsg)

	return req, nil, nil
}

// PostHook is called after a response is received - FULLY ASYNC, NO DATABASE I/O
func (p *LoggerPlugin) PostHook(ctx *context.Context, result *schemas.BifrostResponse, err *schemas.BifrostError) (*schemas.BifrostResponse, *schemas.BifrostError, error) {
	if ctx == nil {
		// Log error but don't fail the request
		p.logger.Error(fmt.Errorf("context is nil in PostHook"))
		return result, err, nil
	}

	// Check if the create operation was dropped - if so, skip the update
	if dropped, ok := (*ctx).Value(DroppedCreateContextKey).(bool); ok && dropped {
		// Create was dropped, skip update to avoid wasted processing and errors
		return result, err, nil
	}

	// Extract request ID from context
	requestID, ok := (*ctx).Value(ContextKey("request-id")).(string)
	if !ok || requestID == "" {
		// Log error but don't fail the request
		p.logger.Error(fmt.Errorf("request-id not found in context or is empty"))
		return result, err, nil
	}

	// Queue the log update message (non-blocking)
	logMsg := p.getLogMessage()
	logMsg.Operation = LogOperationUpdate
	logMsg.RequestID = requestID
	logMsg.Timestamp = time.Now()

	// Prepare update data (latency will be calculated in background worker)
	updateData := &UpdateLogData{}

	if err != nil {
		// Error case
		updateData.Status = "error"
		updateData.ErrorDetails = err
	} else if result != nil {
		// Success case
		updateData.Status = "success"

		// Update model if different from request
		if result.Model != "" {
			updateData.Model = result.Model
		}

		// Update object type if available
		if result.Object != "" {
			updateData.Object = result.Object
		}

		// Token usage
		if result.Usage.TotalTokens > 0 {
			updateData.TokenUsage = &result.Usage
		}

		// Output message and tool calls
		if len(result.Choices) > 0 {
			updateData.OutputMessage = &result.Choices[0].Message

			// Extract tool calls if present
			if result.Choices[0].Message.AssistantMessage != nil &&
				result.Choices[0].Message.AssistantMessage.ToolCalls != nil {
				updateData.ToolCalls = result.Choices[0].Message.AssistantMessage.ToolCalls
			}
		}
	}

	logMsg.UpdateData = updateData

	go func(logMsg *LogMessage) {
		defer p.putLogMessage(logMsg) // Return to pool when done
		if err := p.updateLogEntry(logMsg.RequestID, logMsg.Timestamp, logMsg.UpdateData); err != nil {
			p.logger.Error(fmt.Errorf("failed to update log entry for request %s: %w", logMsg.RequestID, err))
		} else {
			// Call callback if set (for real-time updates)
			p.mu.Lock()
			if p.logCallback != nil {
				// Get the updated log entry for callback
				if updatedEntry, getErr := p.getLogEntry(logMsg.RequestID); getErr == nil {
					p.logCallback(updatedEntry)
				}
			}
			p.mu.Unlock()
		}
	}(logMsg)

	// p.putLogMessage(logMsg)

	return result, err, nil
}

// Cleanup is called when the plugin is being shut down
func (p *LoggerPlugin) Cleanup() error {
	// Stop the cleanup ticker
	if p.cleanupTicker != nil {
		p.cleanupTicker.Stop()
	}

	// Signal the background worker to stop
	close(p.done)

	// Wait for the background worker to finish processing remaining items
	p.wg.Wait()

	// Close the database
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}
