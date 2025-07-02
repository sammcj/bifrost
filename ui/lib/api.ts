import axios, { AxiosInstance, isAxiosError } from "axios";
import {
	ListProvidersResponse,
	ProviderResponse,
	CoreConfig,
	AddProviderRequest,
	UpdateProviderRequest,
	BifrostErrorResponse,
} from "@/lib/types/config";
import { MCPClient, CreateMCPClientRequest, UpdateMCPClientRequest } from "@/lib/types/mcp";
import { LogEntry, LogFilters, LogStats, Pagination } from "./types/logs";

type ServiceResponse<T> = Promise<[T | null, string | null]>;

class ApiService {
	private client: AxiosInstance;

	constructor() {
		let baseURL: string;

		if (process.env.NODE_ENV === "development") {
			// Development mode: Next.js dev server runs on different port than Go server
			const bifrostPort = process.env.NEXT_PUBLIC_BIFROST_PORT || "8080";
			baseURL = `http://localhost:${bifrostPort}/api`;
		} else {
			// Production mode: UI is served by the same Go server, use relative URLs
			// This automatically works with any host:port combination
			baseURL = "/api";
		}

		this.client = axios.create({
			baseURL,
			headers: {
				"Content-Type": "application/json",
			},
		});
	}

	private getErrorMessage(error: unknown): string {
		if (isAxiosError(error) && error.response) {
			const errorData = error.response.data as BifrostErrorResponse;
			if (errorData.error && errorData.error.message) {
				return errorData.error.message;
			}
		}
		if (error instanceof Error) {
			return error.message || "An unexpected error occurred.";
		}
		return "An unexpected error occurred.";
	}

	async getLogs(
		filters: LogFilters,
		pagination: Pagination,
	): ServiceResponse<{
		logs: LogEntry[];
		pagination: Pagination;
		stats: LogStats;
	}> {
		try {
			const params: Record<string, string | number> = {
				limit: pagination.limit,
				offset: pagination.offset,
				sort_by: pagination.sort_by,
				order: pagination.order,
			};

			// Add filters to params if they exist
			if (filters.providers && filters.providers.length > 0) {
				params.providers = filters.providers.join(",");
			}
			if (filters.models && filters.models.length > 0) {
				params.models = filters.models.join(",");
			}
			if (filters.status && filters.status.length > 0) {
				params.status = filters.status.join(",");
			}
			if (filters.objects && filters.objects.length > 0) {
				params.objects = filters.objects.join(",");
			}
			if (filters.start_time) params.start_time = filters.start_time;
			if (filters.end_time) params.end_time = filters.end_time;
			if (filters.min_latency) params.min_latency = filters.min_latency;
			if (filters.max_latency) params.max_latency = filters.max_latency;
			if (filters.min_tokens) params.min_tokens = filters.min_tokens;
			if (filters.max_tokens) params.max_tokens = filters.max_tokens;
			if (filters.content_search) params.content_search = filters.content_search;

			const response = await this.client.get("/logs", { params });
			return [response.data, null];
		} catch (error) {
			return [null, this.getErrorMessage(error)];
		}
	}

	// Provider endpoints
	async getProviders(): ServiceResponse<ListProvidersResponse> {
		try {
			const response = await this.client.get("/providers");
			return [response.data, null];
		} catch (error) {
			return [null, this.getErrorMessage(error)];
		}
	}

	async getProvider(provider: string): ServiceResponse<ProviderResponse> {
		try {
			const response = await this.client.get(`/providers/${provider}`);
			return [response.data, null];
		} catch (error) {
			return [null, this.getErrorMessage(error)];
		}
	}

	async createProvider(data: AddProviderRequest): ServiceResponse<ProviderResponse> {
		try {
			const response = await this.client.post("/providers", data);
			return [response.data, null];
		} catch (error) {
			return [null, this.getErrorMessage(error)];
		}
	}

	async updateProvider(provider: string, data: UpdateProviderRequest): ServiceResponse<ProviderResponse> {
		try {
			const response = await this.client.put(`/providers/${provider}`, data);
			return [response.data, null];
		} catch (error) {
			return [null, this.getErrorMessage(error)];
		}
	}

	async deleteProvider(provider: string): ServiceResponse<ProviderResponse> {
		try {
			const response = await this.client.delete(`/providers/${provider}`);
			return [response.data, null];
		} catch (error) {
			return [null, this.getErrorMessage(error)];
		}
	}

	// MCP endpoints
	async getMCPClients(): ServiceResponse<MCPClient[]> {
		try {
			const response = await this.client.get("/mcp/clients");
			return [response.data, null];
		} catch (error) {
			return [null, this.getErrorMessage(error)];
		}
	}

	async createMCPClient(data: CreateMCPClientRequest): ServiceResponse<null> {
		try {
			await this.client.post("/mcp/client", data);
			return [null, null];
		} catch (error) {
			return [null, this.getErrorMessage(error)];
		}
	}

	async updateMCPClient(name: string, data: UpdateMCPClientRequest): ServiceResponse<null> {
		try {
			await this.client.put(`/mcp/client/${name}`, data);
			return [null, null];
		} catch (error) {
			return [null, this.getErrorMessage(error)];
		}
	}

	async deleteMCPClient(name: string): ServiceResponse<null> {
		try {
			await this.client.delete(`/mcp/client/${name}`);
			return [null, null];
		} catch (error) {
			return [null, this.getErrorMessage(error)];
		}
	}

	async reconnectMCPClient(name: string): ServiceResponse<null> {
		try {
			await this.client.post(`/mcp/client/${name}/reconnect`);
			return [null, null];
		} catch (error) {
			return [null, this.getErrorMessage(error)];
		}
	}

	// Config endpoints
	async saveConfig(): ServiceResponse<null> {
		try {
			await this.client.post("/config/save");
			return [null, null];
		} catch (error) {
			return [null, this.getErrorMessage(error)];
		}
	}

	async getCoreConfig(): ServiceResponse<CoreConfig> {
		try {
			const response = await this.client.get("/config");
			return [response.data, null];
		} catch (error) {
			return [null, this.getErrorMessage(error)];
		}
	}

	async updateCoreConfig(data: CoreConfig): ServiceResponse<null> {
		try {
			await this.client.put("/config", data);
			return [null, null];
		} catch (error) {
			return [null, this.getErrorMessage(error)];
		}
	}
}

export const apiService = new ApiService();
