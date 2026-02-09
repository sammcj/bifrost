import { resolve } from 'path'
import { MCPClientConfig, EnvVarLike } from './pages/mcp-registry.page'

/** Normalize header value from env (string or EnvVarLike) to EnvVarLike */
function toEnvVarLike(v: string | EnvVarLike): EnvVarLike {
  if (typeof v === 'object' && v !== null && 'value' in v) return v as EnvVarLike
  return { value: String(v), env_var: '', from_env: false }
}

/** Build headers from MCP_SSE_HEADERS JSON (injected in workflow). No defaults in code. */
function getSSEHeadersFromEnv(): Record<string, EnvVarLike> {
  const raw = process.env.MCP_SSE_HEADERS
  if (!raw) return {}
  try {
    const parsed = JSON.parse(raw) as Record<string, string | EnvVarLike>
    const out: Record<string, EnvVarLike> = {}
    for (const [k, v] of Object.entries(parsed)) {
      if (v !== undefined && v !== null) out[k] = toEnvVarLike(v)
    }
    return out
  } catch {
    return {}
  }
}

/**
 * Create basic MCP client data
 */
export function createMCPClientData(overrides: Partial<MCPClientConfig> = {}): MCPClientConfig {
  return {
    name: `test_client_${Date.now()}`,
    connectionType: 'http',
    connectionUrl: 'http://localhost:3001/',
    ...overrides,
  }
}

/**
 * Create HTTP MCP client data
 * Uses http-no-ping-server from examples/mcps
 */
export function createHTTPClientData(overrides: Partial<MCPClientConfig> = {}): MCPClientConfig {
  return createMCPClientData({
    connectionType: 'http',
    connectionUrl: 'http://localhost:3001/', // http-no-ping-server
    authType: 'none',
    isPingAvailable: false, // http-no-ping-server doesn't support ping
    ...overrides,
  })
}

/**
 * Create SSE MCP client data.
 * URL and headers are injected via workflow env: MCP_SSE_URL, MCP_SSE_HEADERS (JSON).
 * When unset, uses local http://localhost:3001/sse and no headers (no secrets in code).
 */
export function createSSEClientData(overrides: Partial<MCPClientConfig> = {}): MCPClientConfig {
  const connectionUrl = process.env.MCP_SSE_URL ?? 'http://localhost:3001/sse'
  const headers = getSSEHeadersFromEnv()
  const hasHeaders = headers && Object.keys(headers).length > 0
  return createMCPClientData({
    connectionType: 'sse',
    connectionUrl,
    authType: hasHeaders ? 'headers' : 'none',
    headers: hasHeaders ? headers : undefined,
    isPingAvailable: false,
    ...overrides,
  })
}

/**
 * Create STDIO MCP client data
 * Uses test-tools-server from examples/mcps
 */
export function createSTDIOClientData(overrides: Partial<MCPClientConfig> = {}): MCPClientConfig {
  // Use the built test-tools-server
  const serverPath = resolve(__dirname, '../../../../examples/mcps/test-tools-server/dist/index.js')

  return createMCPClientData({
    name: `stdio_client_${Date.now()}`,
    connectionType: 'stdio',
    command: 'node',
    args: serverPath, // Run the actual MCP server
    ...overrides,
  })
}

/**
 * Create HTTP client with headers auth
 */
export function createHTTPClientWithHeaders(overrides: Partial<MCPClientConfig> = {}): MCPClientConfig {
  return createMCPClientData({
    connectionType: 'http',
    connectionUrl: 'http://localhost:3001/', // http-no-ping-server
    authType: 'headers',
    isPingAvailable: false,
    ...overrides,
  })
}

/**
 * Create HTTP client with OAuth auth (minimal config for testing)
 * Note: http-no-ping-server doesn't support OAuth, so this is for UI testing only
 */
export function createHTTPClientWithOAuth(overrides: Partial<MCPClientConfig> = {}): MCPClientConfig {
  return createMCPClientData({
    name: `oauth_client_${Date.now()}`,
    connectionType: 'http',
    connectionUrl: 'http://localhost:3001/', // http-no-ping-server
    authType: 'oauth',
    oauthClientId: 'test-client-id',
    oauthAuthorizeUrl: 'http://localhost:3001/oauth/authorize',
    oauthTokenUrl: 'http://localhost:3001/oauth/token',
    isPingAvailable: false,
    ...overrides,
  })
}

/**
 * Create client with code mode enabled
 */
export function createCodeModeClientData(overrides: Partial<MCPClientConfig> = {}): MCPClientConfig {
  return createMCPClientData({
    isCodeMode: true,
    ...overrides,
  })
}
