import { BifrostConfig, CacheConfig, CoreConfig } from "@/lib/types/config";
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

		// Update core configuration
		updateCoreConfig: builder.mutation<null, CoreConfig>({
			query: (data) => ({
				url: "/config",
				method: "PUT",
				body: data,
			}),
			invalidatesTags: ["Config"],
		}),

		// Get cache configuration
		getCacheConfig: builder.query<CacheConfig, void>({
			query: () => "/config/cache",
			providesTags: ["CacheConfig"],
		}),

		// Update cache configuration
		updateCacheConfig: builder.mutation<{ config: CacheConfig }, CacheConfig>({
			query: (data) => ({
				url: "/config/cache",
				method: "PUT",
				body: data,
			}),
			invalidatesTags: ["CacheConfig"],
		}),
	}),
});

export const {
	useGetCoreConfigQuery,
	useUpdateCoreConfigMutation,
	useGetCacheConfigQuery,
	useUpdateCacheConfigMutation,
	useLazyGetCoreConfigQuery,
	useLazyGetCacheConfigQuery,
} = configApi;
