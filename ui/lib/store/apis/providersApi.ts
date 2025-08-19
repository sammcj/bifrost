import { AddProviderRequest, ListProvidersResponse, ProviderResponse, UpdateProviderRequest } from "@/lib/types/config";
import { DBKey } from "@/lib/types/governance";
import { baseApi } from "./baseApi";

export const providersApi = baseApi.injectEndpoints({
	endpoints: (builder) => ({
		// Get all providers
		getProviders: builder.query<ListProvidersResponse, void>({
			query: () => "/providers",
			providesTags: ["Providers"],
		}),

		// Get single provider
		getProvider: builder.query<ProviderResponse, string>({
			query: (provider) => `/providers/${provider}`,
			providesTags: (result, error, provider) => [{ type: "Providers", id: provider }],
		}),

		// Create new provider
		createProvider: builder.mutation<ProviderResponse, AddProviderRequest>({
			query: (data) => ({
				url: "/providers",
				method: "POST",
				body: data,
			}),
			invalidatesTags: ["Providers"],
		}),

		// Update existing provider
		updateProvider: builder.mutation<ProviderResponse, { provider: string; data: UpdateProviderRequest }>({
			query: ({ provider, data }) => ({
				url: `/providers/${provider}`,
				method: "PUT",
				body: data,
			}),
			invalidatesTags: (result, error, { provider }) => ["Providers", { type: "Providers", id: provider }],
		}),

		// Delete provider
		deleteProvider: builder.mutation<ProviderResponse, string>({
			query: (provider) => ({
				url: `/providers/${provider}`,
				method: "DELETE",
			}),
			invalidatesTags: ["Providers"],
		}),

		// Get all available keys from all providers for governance selection
		getAllKeys: builder.query<DBKey[], void>({
			query: () => "/keys",
			providesTags: ["DBKeys"],
		}),
	}),
});

export const {
	useGetProvidersQuery,
	useGetProviderQuery,
	useCreateProviderMutation,
	useUpdateProviderMutation,
	useDeleteProviderMutation,
	useGetAllKeysQuery,
	useLazyGetProvidersQuery,
	useLazyGetProviderQuery,
	useLazyGetAllKeysQuery,
} = providersApi;
