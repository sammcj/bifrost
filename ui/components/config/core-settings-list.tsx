"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import { CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { AlertTriangle, Loader2 } from "lucide-react";
import { CoreConfig } from "@/lib/types/config";
import { apiService } from "@/lib/api";
import { toast } from "sonner";
import { Separator } from "@radix-ui/react-separator";
import { Alert, AlertDescription } from "../ui/alert";
import { Input } from "../ui/input";
import { Textarea } from "../ui/textarea";
import { parseArrayFromText } from "@/lib/utils/array";

export default function CoreSettingsList() {
	const [config, setConfig] = useState<CoreConfig>({
		drop_excess_requests: false,
		initial_pool_size: 300,
	});
	const [isLoading, setIsLoading] = useState(true);
	const [localValues, setLocalValues] = useState<{
		initial_pool_size: string;
		prometheus_labels: string;
	}>({
		initial_pool_size: "300",
		prometheus_labels: "",
	});

	// Use refs to store timeout IDs
	const poolSizeTimeoutRef = useRef<NodeJS.Timeout | undefined>(undefined);
	const prometheusLabelsTimeoutRef = useRef<NodeJS.Timeout | undefined>(undefined);

	useEffect(() => {
		const fetchConfig = async () => {
			const [coreConfig, error] = await apiService.getCoreConfig();
			if (error) {
				toast.error(error);
			} else if (coreConfig) {
				setConfig(coreConfig);
				setLocalValues({
					initial_pool_size: coreConfig.initial_pool_size?.toString() || "300",
					prometheus_labels: coreConfig.prometheus_labels || "",
				});
			}
			setIsLoading(false);
		};
		fetchConfig();
	}, []);

	const updateConfig = useCallback(
		async (field: keyof CoreConfig, value: boolean | number | string[]) => {
			const newConfig = { ...config, [field]: value };
			setConfig(newConfig);

			const [, error] = await apiService.updateCoreConfig(newConfig);
			if (error) {
				toast.error(error);
			} else {
				toast.success("Core setting updated successfully.");
			}
		},
		[config],
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
				const numValue = parseInt(value);
				if (!isNaN(numValue) && numValue > 0) {
					updateConfig("initial_pool_size", numValue);
				}
			}, 1000);
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
		};
	}, []);

	if (isLoading) {
		return (
			<div className="flex h-64 items-center justify-center">
				<Loader2 className="h-4 w-4 animate-spin" />
			</div>
		);
	}

	return (
		<div>
			<CardHeader className="mb-4 px-0">
				<CardTitle className="flex items-center gap-2">Core System Settings</CardTitle>
				<CardDescription>Configure core Bifrost settings like request handling, pool sizes, and system behavior.</CardDescription>
			</CardHeader>
			<div className="space-y-6">
				{/* Drop Excess Requests */}
				<div className="flex items-center justify-between space-x-2 rounded-lg border p-4">
					<div className="space-y-0.5">
						<label className="text-sm font-medium">Drop Excess Requests</label>
						<p className="text-muted-foreground text-sm">If enabled, Bifrost will drop requests that exceed pool capacity.</p>
					</div>
					<Switch
						checked={config.drop_excess_requests}
						onCheckedChange={(checked) => handleConfigChange("drop_excess_requests", checked)}
					/>
				</div>

				<Separator />

				<Alert>
					<AlertTriangle className="h-4 w-4" />
					<AlertDescription>
						The settings below won't affect current connections. You will need to save the configuration for changes to take effect on the
						next restart.
					</AlertDescription>
				</Alert>

				<div className="flex items-center justify-between space-x-2 rounded-lg border p-4">
					<div className="space-y-0.5">
						<label className="text-sm font-medium">Initial Pool Size</label>
						<p className="text-muted-foreground text-sm">The initial connection pool size.</p>
					</div>
					<Input
						type="number"
						className="w-24"
						value={localValues.initial_pool_size}
						onChange={(e) => handlePoolSizeChange(e.target.value)}
						min="1"
					/>
				</div>

				<div className="space-y-2 rounded-lg border p-4">
					<div className="space-y-0.5">
						<label className="text-sm font-medium">Prometheus Labels</label>
						<p className="text-muted-foreground text-sm">Comma-separated list of custom labels to add to the Prometheus metrics.</p>
					</div>
					<Textarea
						className="h-24"
						placeholder="teamId, projectId, environment"
						value={localValues.prometheus_labels}
						onChange={(e) => handlePrometheusLabelsChange(e.target.value)}
					/>
				</div>
			</div>
		</div>
	);
}
