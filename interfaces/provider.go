package interfaces

// LLMUsage represents token usage information
type LLMUsage struct {
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	Latency          float64 `json:"latency,omitempty"`
}

// LLMInteractionCost represents cost information for LLM interactions
type LLMInteractionCost struct {
	Input  float64 `json:"input"`
	Output float64 `json:"output"`
	Total  float64 `json:"total"`
}

// Function represents a function definition for tool calls
type Function struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

// Tool represents a tool that can be used with the model
type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// ModelParameters represents the parameters for model requests
type ModelParameters struct {
	TestRunEntryID *string     `json:"testRunEntryId,omitempty"`
	PromptTools    *[]string   `json:"promptTools,omitempty"`
	ToolChoice     *string     `json:"toolChoice,omitempty"`
	Tools          *[]Tool     `json:"tools,omitempty"`
	FunctionCall   *string     `json:"functionCall,omitempty"`
	Functions      *[]Function `json:"functions,omitempty"`
	// Dynamic parameters
	ExtraParams map[string]interface{} `json:"-"`
}

// RequestOptions represents options for model requests
type RequestOptions struct {
	UseCache       bool   `json:"useCache,omitempty"`
	WaitForModel   bool   `json:"waitForModel,omitempty"`
	CompletionType string `json:"CompletionType,omitempty"`
}

// FunctionCall represents a function call in a tool call
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolCall represents a tool call in a message
type ToolCall struct {
	Type     string       `json:"type"`
	ID       string       `json:"id"`
	Function FunctionCall `json:"function"`
}

// ModelChatMessageRole represents the role of a chat message
type ModelChatMessageRole string

const (
	RoleAssistant ModelChatMessageRole = "assistant"
	RoleUser      ModelChatMessageRole = "user"
	RoleSystem    ModelChatMessageRole = "system"
	RoleModel     ModelChatMessageRole = "model"
	RoleChatbot   ModelChatMessageRole = "chatbot"
	RoleTool      ModelChatMessageRole = "tool"
)

// CompletionResponseChoice represents a choice in the completion response
type CompletionResponseChoice struct {
	Role         ModelChatMessageRole `json:"role"`
	Content      string               `json:"content"`
	FunctionCall *FunctionCall        `json:"function_call,omitempty"`
	ToolCalls    []ToolCall           `json:"tool_calls,omitempty"`
}

// CompletionResultChoice represents a choice in the completion result
type CompletionResultChoice struct {
	Index        int                      `json:"index"`
	Message      CompletionResponseChoice `json:"message"`
	FinishReason string                   `json:"finish_reason,omitempty"`
	LogProbs     interface{}              `json:"logprobs,omitempty"`
}

// ToolResult represents the result of a tool call
type ToolResult struct {
	Role       ModelChatMessageRole `json:"role"`
	Content    string               `json:"content"`
	ToolCallID string               `json:"tool_call_id"`
}

// ToolCallResult represents a single tool call result
type ToolCallResult struct {
	Name   string      `json:"name"`
	Result interface{} `json:"result"`
	Type   string      `json:"type"`
	ID     string      `json:"id"`
}

// ToolCallResults represents a collection of tool call results
type ToolCallResults struct {
	Version int              `json:"version"`
	Results []ToolCallResult `json:"results"`
}

// CompletionResult represents the complete result from a model completion
type CompletionResult struct {
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
	ID              string                   `json:"id"`
	Choices         []CompletionResultChoice `json:"choices"`
	ToolCallResult  interface{}              `json:"tool_call_result,omitempty"`
	ToolCallResults *ToolCallResults         `json:"toolCallResults,omitempty"`
	Provider        SupportedModelProvider   `json:"provider,omitempty"`
	Usage           LLMUsage                 `json:"usage"`
	Cost            *LLMInteractionCost      `json:"cost,omitempty"`
	Model           string                   `json:"model,omitempty"`
	Created         string                   `json:"created,omitempty"`
	ModelParams     interface{}              `json:"modelParams,omitempty"`
	Trace           *struct {
		Input  interface{} `json:"input"`
		Output interface{} `json:"output,omitempty"`
	} `json:"trace,omitempty"`
	RetrievedContext        interface{}            `json:"retrievedContext,omitempty"`
	VariableBoundRetrievals map[string]interface{} `json:"variableBoundRetrievals,omitempty"`
}

type SupportedModelProvider string

const (
	OpenAI      SupportedModelProvider = "openai"
	Azure       SupportedModelProvider = "azure"
	HuggingFace SupportedModelProvider = "huggingface"
	Anthropic   SupportedModelProvider = "anthropic"
	Google      SupportedModelProvider = "google"
	Groq        SupportedModelProvider = "groq"
	Bedrock     SupportedModelProvider = "bedrock"
	Maxim       SupportedModelProvider = "maxim"
	Cohere      SupportedModelProvider = "cohere"
	Ollama      SupportedModelProvider = "ollama"
	Lmstudio    SupportedModelProvider = "lmstudio"
)

// Provider defines the interface for AI model providers
type Provider interface {
	GetProviderKey() SupportedModelProvider
	GetConfig() interface{}
	TextCompletion(model, key, text string, params *ModelParameters) (*CompletionResult, error)
	ChatCompletion(model, key string, messages []interface{}, params *ModelParameters) (*CompletionResult, error)
}
