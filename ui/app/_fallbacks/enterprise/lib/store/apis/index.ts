// Placeholder for enterprise APIs
// Export empty objects when enterprise features are not available

export const scimApi = null
export const guardrailsApi = null
export const clusterApi = null

// Fallback hooks that return no data for non-enterprise builds
export const useGetUserInfoQuery = (_arg?: any, _options?: any) => ({ 
  data: undefined, 
  isLoading: false, 
  error: undefined 
})

export const useLazyGetUserInfoQuery = () => [
  () => Promise.resolve({ data: undefined }),
  { 
    data: undefined, 
    isLoading: false, 
    error: undefined 
  }
]

export const useLogoutMutation = () => [
  () => Promise.resolve(),
  { 
    isLoading: false, 
    error: undefined 
  }
]

export const useGetClusterNodesQuery = (_arg?: any, _options?: any) => ({
  data: [],
  isLoading: false,
  error: undefined
})

// Empty apis array when enterprise features are not available
export const apis = []

