import { BifrostConfig, CoreConfig, LatestReleaseResponse } from "@/lib/types/config";
import axios from "axios";
import { baseApi } from "./baseApi";

export const configApi = baseApi.injectEndpoints({
	endpoints: (builder) => ({
		// Get core configuration
		getCoreConfig: builder.query<BifrostConfig, { fromDB?: boolean }>({
			query: ({ fromDB = false } = {}) => ({
				url: "/config",
				params: { from_db: fromDB },
			}),
			providesTags: ["Config"],
		}),

		// Get version information
		getVersion: builder.query<string, void>({
			query: () => ({
				url: "/version",
			}),
		}),

	// Get latest release from public site
	getLatestRelease: builder.query<LatestReleaseResponse, void>({
		queryFn: async () => {
			try {
				const response = await axios.get('https://getbifrost.ai/latest-release', {
					timeout: 3000, // 3 second timeout
					headers: {
						Accept: 'application/json',						
					},
					maxRedirects: 5,
					validateStatus: (status) => status >= 200 && status < 300,
				})
				return { data: response.data }
			} catch (error) {
				if (axios.isAxiosError(error)) {
					if (error.code === 'ECONNABORTED' || error.code === 'ETIMEDOUT') {
						console.warn('Latest release fetch timed out after 3s')
						return { error: { status: 'TIMEOUT_ERROR', error: 'Request timeout', data: { error: { message: 'Request timeout' } } } }
					}
					console.error('Latest release fetch error:', error.message)
				} else {
					console.error('Latest release fetch error:', error)
				}
				return { error: { status: 'FETCH_ERROR', error: String(error), data: { error: { message: 'Network error' } } } }
			}
		},
		keepUnusedDataFor: 300 * 1000, // Cache for 5 minutes
	}),

		// Update core configuration
		updateCoreConfig: builder.mutation<null, CoreConfig>({
			query: (data) => ({
				url: "/config",
				method: "PUT",
				body: data,
			}),
			invalidatesTags: ["Config"],
		}),
	}),
});

export const {
	useGetVersionQuery,
	useGetCoreConfigQuery,
	useUpdateCoreConfigMutation,
	useLazyGetCoreConfigQuery,
	useGetLatestReleaseQuery,
	useLazyGetLatestReleaseQuery,
} = configApi;
