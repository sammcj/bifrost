"use client";

import { useState, useEffect } from "react";
import Header from "@/components/header";
import { CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Settings, Database, Zap, Save, RefreshCw } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { ProviderResponse } from "@/lib/types/config";
import { apiService } from "@/lib/api";
import CoreSettingsList from "@/components/config/core-settings-list";
import ProvidersList from "@/components/config/providers-list";
import MCPClientsList from "@/components/config/mcp-clients-lists";
import { MCPClient } from "@/lib/types/mcp";
import FullPageLoader from "@/components/full-page-loader";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";

export default function ConfigPage() {
	const [activeTab, setActiveTab] = useState("providers");
	const [isLoadingProviders, setIsLoadingProviders] = useState(true);
	const [isLoadingMcpClients, setIsLoadingMcpClients] = useState(true);
	const [providers, setProviders] = useState<ProviderResponse[]>([]);
	const [mcpClients, setMcpClients] = useState<MCPClient[]>([]);

	const { toast } = useToast();

	// Load configuration data
	useEffect(() => {
		loadProviders();
		loadMcpClients();
	}, []);

	const loadProviders = async () => {
		const [data, error] = await apiService.getProviders();
		setIsLoadingProviders(false);

		if (error) {
			toast({
				title: "Error",
				description: error,
				variant: "destructive",
			});
			return;
		}
		setProviders(data?.providers || []);
	};

	const loadMcpClients = async () => {
		const [data, error] = await apiService.getMCPClients();
		setIsLoadingMcpClients(false);

		if (error) {
			toast({
				title: "Error",
				description: error,
				variant: "destructive",
			});
			return;
		}

		setMcpClients(data || []);
	};

	const handleSaveConfig = async () => {
		const [, error] = await apiService.saveConfig();

		if (error) {
			toast({
				title: "Error",
				description: error,
				variant: "destructive",
			});
		} else {
			toast({
				title: "Success",
				description: "Configuration saved successfully",
			});
		}
	};

	return (
		<div className="bg-background">
			<Header title="Configuration" />
			{isLoadingProviders || isLoadingMcpClients ? (
				<FullPageLoader />
			) : (
				<div className="space-y-6">
					{/* Page Header */}
					<div className="flex items-center justify-between">
						<div>
							<h1 className="text-3xl font-bold">Configuration</h1>
							<p className="text-muted-foreground mt-2">Configure AI providers, API keys, and system settings for your Bifrost instance.</p>
						</div>

						<TooltipProvider>
							<Tooltip>
								<TooltipTrigger asChild>
									<Button onClick={handleSaveConfig} variant="outline" disabled={isLoadingProviders || isLoadingMcpClients}>
										<Save className="h-4 w-4" />
										Save Config
									</Button>
								</TooltipTrigger>
								<TooltipContent>Persist configuration for next server start.</TooltipContent>
							</Tooltip>
						</TooltipProvider>
					</div>

					{/* Configuration Tabs */}
					<Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
						<TabsList className="grid h-12 w-full grid-cols-3">
							<TabsTrigger value="providers" className="flex items-center gap-2">
								<Database className="h-4 w-4" />
								Providers
								<Badge variant="default" className="ml-1">
									{providers.length}
								</Badge>
							</TabsTrigger>
							<TabsTrigger value="mcp" className="flex items-center gap-2">
								<Zap className="h-4 w-4" />
								MCP Clients
								{mcpClients.length > 0 && (
									<Badge variant="default" className="ml-1">
										{mcpClients.length}
									</Badge>
								)}
							</TabsTrigger>
							<TabsTrigger value="core" className="flex items-center gap-2">
								<Settings className="h-4 w-4" />
								Core Settings
							</TabsTrigger>
						</TabsList>

						{/* Providers Tab */}
						<TabsContent value="providers" className="space-y-4">
							<ProvidersList providers={providers} onRefresh={loadProviders} />
						</TabsContent>

						{/* MCP Tools Tab */}
						<TabsContent value="mcp" className="space-y-4">
							<MCPClientsList />
						</TabsContent>

						{/* Core Settings Tab */}
						<TabsContent value="core" className="space-y-4">
							<CoreSettingsList />
						</TabsContent>
					</Tabs>
				</div>
			)}
		</div>
	);
}
