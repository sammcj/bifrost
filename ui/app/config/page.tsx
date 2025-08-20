"use client";

import CacheConfigForm from "@/app/config/views/cache-config-form";
import FullPageLoader from "@/components/full-page-loader";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import { getErrorMessage, useGetCoreConfigQuery, useGetDroppedRequestsQuery, useUpdateCoreConfigMutation } from "@/lib/store";
import { CoreConfig } from "@/lib/types/config";
import { parseArrayFromText } from "@/lib/utils/array";
import { validateOrigins } from "@/lib/utils/validation";
import { AlertTriangle } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import { toast } from "sonner";

const defaultConfig = {
	drop_excess_requests: false,
	initial_pool_size: 300,
	prometheus_labels: [],
	enable_logging: true,
	enable_governance: true,
	enforce_governance_header: false,
	allow_direct_keys: false,
	enable_caching: false,
	allowed_origins: [],
};

export default function ConfigPage() {
	const [droppedRequests, setDroppedRequests] = useState<number>(0);
	// RTK Query hooks
	const { data: droppedRequestsData } = useGetDroppedRequestsQuery();
	const { data: bifrostConfig, isLoading } = useGetCoreConfigQuery({ fromDB: true });
	const config = bifrostConfig?.client_config;
	const [needsRestart, setNeedsRestart] = useState<boolean>(false);
	const [updateCoreConfig] = useUpdateCoreConfigMutation();

	const [localValues, setLocalValues] = useState<{
		initial_pool_size: string;
		prometheus_labels: string;
		allowed_origins: string;
	}>({
		initial_pool_size: "300",
		prometheus_labels: "",
		allowed_origins: "",
	});

	// Handle dropped requests data from RTK Query
	useEffect(() => {
		if (droppedRequestsData) {
			setDroppedRequests(droppedRequestsData.dropped_requests);
		}
	}, [droppedRequestsData]);

	// Use refs to store timeout IDs
	const poolSizeTimeoutRef = useRef<NodeJS.Timeout | undefined>(undefined);
	const prometheusLabelsTimeoutRef = useRef<NodeJS.Timeout | undefined>(undefined);
	const allowedOriginsTimeoutRef = useRef<NodeJS.Timeout | undefined>(undefined);

	// Update local values when config is loaded
	useEffect(() => {
		if (bifrostConfig && config) {
			setLocalValues({
				initial_pool_size: config?.initial_pool_size?.toString() || "300",
				prometheus_labels: config?.prometheus_labels?.join(", ") || "",
				allowed_origins: config?.allowed_origins?.join(", ") || "",
			});
		}
	}, [config]);

	const updateConfig = useCallback(
		async (field: keyof CoreConfig, value: boolean | number | string[]) => {
			try {
				await updateCoreConfig({ ...(config ?? defaultConfig), [field]: value }).unwrap();
				toast.success("Core setting updated successfully.");
			} catch (error) {
				toast.error(getErrorMessage(error));
			}
		},
		[config, updateCoreConfig],
	);

	const handleConfigChange = async (field: keyof CoreConfig, value: boolean | number | string[]) => {
		await updateConfig(field, value);
	};

	const handlePoolSizeChange = useCallback(
		(value: string) => {
			setLocalValues((prev) => ({ ...prev, initial_pool_size: value }));

			// Clear existing timeout
			if (poolSizeTimeoutRef.current) {
				clearTimeout(poolSizeTimeoutRef.current);
			}

			// Set new timeout
			poolSizeTimeoutRef.current = setTimeout(() => {
				const numValue = Number.parseInt(value);
				if (!isNaN(numValue) && numValue > 0) {
					updateConfig("initial_pool_size", numValue);
				}
			}, 1000);
			setNeedsRestart(true);
		},
		[updateConfig],
	);

	const handlePrometheusLabelsChange = useCallback(
		(value: string) => {
			setLocalValues((prev) => ({ ...prev, prometheus_labels: value }));

			// Clear existing timeout
			if (prometheusLabelsTimeoutRef.current) {
				clearTimeout(prometheusLabelsTimeoutRef.current);
			}

			// Set new timeout
			prometheusLabelsTimeoutRef.current = setTimeout(() => {
				updateConfig("prometheus_labels", parseArrayFromText(value));
			}, 1000);
			setNeedsRestart(true);
		},
		[updateConfig],
	);

	const handleAllowedOriginsChange = useCallback(
		(value: string) => {
			setLocalValues((prev) => ({ ...prev, allowed_origins: value }));

			// Clear existing timeout
			if (allowedOriginsTimeoutRef.current) {
				clearTimeout(allowedOriginsTimeoutRef.current);
			}

			// Set new timeout
			allowedOriginsTimeoutRef.current = setTimeout(() => {
				const origins = parseArrayFromText(value);
				const validation = validateOrigins(origins);

				if (validation.isValid || origins.length === 0) {
					updateConfig("allowed_origins", origins);
				} else {
					toast.error(`Invalid origins: ${validation.invalidOrigins.join(", ")}. Origins must be valid URLs like https://example.com`);
				}
			}, 1000);
			setNeedsRestart(true);
		},
		[updateConfig],
	);

	// Cleanup timeouts on unmount
	useEffect(() => {
		return () => {
			if (poolSizeTimeoutRef.current) {
				clearTimeout(poolSizeTimeoutRef.current);
			}
			if (prometheusLabelsTimeoutRef.current) {
				clearTimeout(prometheusLabelsTimeoutRef.current);
			}
			if (allowedOriginsTimeoutRef.current) {
				clearTimeout(allowedOriginsTimeoutRef.current);
			}
		};
	}, []);

	return isLoading ? (
		<FullPageLoader />
	) : (
		<div className="dark:bg-card space-y-6 bg-white">
			{/* Page Header */}
			<div>
				<CardHeader className="mb-4 px-0">
					<CardTitle className="flex items-center gap-2">Core System Settings</CardTitle>
					<CardDescription>Configure core Bifrost settings like request handling, pool sizes, and system behavior.</CardDescription>
				</CardHeader>
				<div className="space-y-6">
					{/* Drop Excess Requests */}
					<div className="flex items-center justify-between space-x-2 rounded-lg border p-4">
						<div className="space-y-0.5">
							<label htmlFor="drop-excess-requests" className="text-sm font-medium">
								Drop Excess Requests
							</label>
							<p className="text-muted-foreground text-sm">
								If enabled, Bifrost will drop requests that exceed pool capacity.{" "}
								{config?.drop_excess_requests && droppedRequests > 0 ? (
									<span>
										Have dropped <b>{droppedRequests} requests</b> since last restart.
									</span>
								) : (
									<></>
								)}
							</p>
						</div>
						<Switch
							id="drop-excess-requests"
							size="md"
							checked={config?.drop_excess_requests}
							onCheckedChange={(checked) => handleConfigChange("drop_excess_requests", checked)}
						/>
					</div>

					{config?.enable_governance && (
						<div className="flex items-center justify-between space-x-2 rounded-lg border p-4">
							<div className="space-y-0.5">
								<label htmlFor="enforce-governance" className="text-sm font-medium">
									Enforce Virtual Keys
								</label>
								<p className="text-muted-foreground text-sm">
									Enforce the use of a virtual key for all requests. If enabled, requests without the <b>x-bf-vk</b> header will be
									rejected.
								</p>
							</div>
							<Switch
								id="enforce-governance"
								size="md"
								checked={config.enforce_governance_header}
								onCheckedChange={(checked) => handleConfigChange("enforce_governance_header", checked)}
							/>
						</div>
					)}

					<div className="flex items-center justify-between space-x-2 rounded-lg border p-4">
						<div className="space-y-0.5">
							<label htmlFor="allow-direct-keys" className="text-sm font-medium">
								Allow Direct API Keys
							</label>
							<p className="text-muted-foreground text-sm">
								Allow API keys to be passed directly in request headers (<b>Authorization</b> or <b>x-api-key</b>). Bifrost will directly
								use the key.
							</p>
						</div>
						<Switch
							id="allow-direct-keys"
							size="md"
							checked={config?.allow_direct_keys}
							onCheckedChange={(checked) => handleConfigChange("allow_direct_keys", checked)}
						/>
					</div>

					<Alert variant="destructive">
						<AlertTriangle className="h-4 w-4" />
						<AlertDescription>
							The settings below require a Bifrost service restart to take effect. Current connections will continue with existing settings
							until restart.
						</AlertDescription>
					</Alert>

					<div>
						<div className="flex items-center justify-between space-x-2 rounded-lg border p-4">
							<div className="space-y-0.5">
								<label htmlFor="initial-pool-size" className="text-sm font-medium">
									Initial Pool Size
								</label>
								<p className="text-muted-foreground text-sm">The initial connection pool size.</p>
							</div>
							<Input
								id="initial-pool-size"
								type="number"
								className="w-24"
								value={localValues.initial_pool_size}
								onChange={(e) => handlePoolSizeChange(e.target.value)}
								min="1"
							/>
						</div>
						{needsRestart && <RestartWarning />}
					</div>

					<div>
						<div className="flex items-center justify-between space-x-2 rounded-lg border p-4">
							<div className="space-y-0.5">
								<label htmlFor="enable-logging" className="text-sm font-medium">
									Enable Logs
								</label>
								<p className="text-muted-foreground text-sm">
									Enable logging of requests and responses to a SQL database. This can add 40-60mb of overhead to the system memory.
								</p>
							</div>
							<Switch
								id="enable-logging"
								size="md"
								checked={config?.enable_logging}
								onCheckedChange={(checked) => handleConfigChange("enable_logging", checked)}
							/>
						</div>
						{needsRestart && <RestartWarning />}
					</div>

					<div>
						<div className="flex items-center justify-between space-x-2 rounded-lg border p-4">
							<div className="space-y-0.5">
								<label htmlFor="enable-governance" className="text-sm font-medium">
									Enable Governance
								</label>
								<p className="text-muted-foreground text-sm">
									Enable governance on requests. You can configure budgets and rate limits in the <b>Governance</b> tab.
								</p>
							</div>
							<Switch
								id="enable-governance"
								size="md"
								checked={config?.enable_governance}
								onCheckedChange={(checked) => handleConfigChange("enable_governance", checked)}
							/>
						</div>
						{needsRestart && <RestartWarning />}
					</div>

					<div>
						<div className="rounded-lg border p-4">
							<div className="flex items-center justify-between space-x-2">
								<div className="space-y-0.5">
									<label htmlFor="enable-caching" className="text-sm font-medium">
										Enable Semantic Caching
									</label>
									<p className="text-muted-foreground text-sm">
										Enable semantic caching for requests. Send <b>x-bf-cache-key</b> header with requests to use semantic caching.
									</p>
								</div>
								<Switch
									id="enable-caching"
									size="md"
									checked={config?.enable_caching}
									onCheckedChange={(checked) => handleConfigChange("enable_caching", checked)}
								/>
							</div>

							{config?.enable_caching && (
								<div className="mt-4 space-y-4">
									<Separator />
									<CacheConfigForm />
								</div>
							)}
						</div>

						{needsRestart && <RestartWarning />}
					</div>

					<div>
						<div className="space-y-2 rounded-lg border p-4">
							<div className="space-y-0.5">
								<label htmlFor="prometheus-labels" className="text-sm font-medium">
									Prometheus Labels
								</label>
								<p className="text-muted-foreground text-sm">Comma-separated list of custom labels to add to the Prometheus metrics.</p>
							</div>
							<Textarea
								id="prometheus-labels"
								className="h-24"
								placeholder="teamId, projectId, environment"
								value={localValues.prometheus_labels}
								onChange={(e) => handlePrometheusLabelsChange(e.target.value)}
							/>
						</div>
						{needsRestart && <RestartWarning />}
					</div>

					<div>
						<div className="space-y-2 rounded-lg border p-4">
							<div className="space-y-0.5">
								<label htmlFor="allowed-origins" className="text-sm font-medium">
									Allowed Origins
								</label>
								<p className="text-muted-foreground text-sm">
									Comma-separated list of allowed origins for CORS and WebSocket connections. Localhost origins are always allowed. Each
									origin must be a complete URL with protocol (e.g., https://app.example.com).
								</p>
							</div>
							<Textarea
								id="allowed-origins"
								className="h-24"
								placeholder="https://app.example.com, https://staging.example.com"
								value={localValues.allowed_origins}
								onChange={(e) => handleAllowedOriginsChange(e.target.value)}
							/>
						</div>
						{needsRestart && <RestartWarning />}
					</div>
				</div>
			</div>
		</div>
	);
}

const RestartWarning = () => {
	return <div className="text-muted-foreground mt-2 pl-4 text-xs font-semibold">Need to restart Bifrost to apply changes.</div>;
};
