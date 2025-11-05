// Fallback OAuth Token Manager for non-enterprise builds
// These functions return null/no-op when enterprise features are not available

export const getAccessToken = (): string | null => null

export const getRefreshToken = (): string | null => null

export const getTokenExpiry = (): number | null => null

export const isTokenExpired = (): boolean => false

export const setOAuthTokens = (accessToken: string, refreshToken?: string, expiresIn?: number) => {
  // No-op in non-enterprise builds
}

export const clearOAuthStorage = () => {
  // No-op in non-enterprise builds
}

export const getRefreshState = () => ({
  isRefreshing: false,
  refreshPromise: null
})

export const setRefreshState = (refreshing: boolean, promise: Promise<any> | null = null) => {
  // No-op in non-enterprise builds
}

export const REFRESH_TOKEN_ENDPOINT = ''

// User info type definition (matching enterprise version)
export interface UserInfo {
  name?: string
  email?: string
  picture?: string
  preferred_username?: string
  given_name?: string
  family_name?: string
}

// Fallback getUserInfo that returns null for non-enterprise builds
export const getUserInfo = (): UserInfo | null => null

// Fallback setUserInfo - no-op
export const setUserInfo = (userInfo: UserInfo) => {
  // No-op in non-enterprise builds
}

// Fallback clearUserInfo - no-op
export const clearUserInfo = () => {
  // No-op in non-enterprise builds
}

