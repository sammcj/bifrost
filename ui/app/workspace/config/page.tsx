"use client";

import FullPageLoader from "@/components/fullPageLoader";
import { IS_ENTERPRISE } from "@/lib/constants/config";
import { useGetCoreConfigQuery } from "@/lib/store";
import APIKeysView from "@enterprise/components/api-keys/APIKeysView";
import { useQueryState } from "nuqs";
import { useEffect } from "react";
import CachingView from "./views/cachingView";
import ClientSettingsView from "./views/clientSettingsView";
import GovernanceView from "./views/governanceView";
import LoggingView from "./views/loggingView";
import ObservabilityView from "./views/observabilityView";
import PerformanceTuningView from "./views/performanceTuningView";
import PricingConfigView from "./views/pricingConfigView";
import ProxyView from "./views/proxyView";
import SecurityView from "./views/securityView";

const validTabs = [
	"client-settings",
	"pricing-config",
	"logging",
	"governance",
	"caching",
	"observability",
	"security",
	"proxy",
	"api-keys",
	"performance-tuning",
];

export default function ConfigPage() {
	const [activeTab, setActiveTab] = useQueryState("tab");
	const { isLoading } = useGetCoreConfigQuery({ fromDB: true });

	// Redirect to default tab if none specified
	useEffect(() => {
		if (!activeTab || !validTabs.includes(activeTab)) {
			setActiveTab("client-settings");
		}
	}, [activeTab, setActiveTab]);

	if (isLoading) {
		return <FullPageLoader />;
	}

	return (
		<div className="mx-auto flex w-full max-w-7xl">
			{activeTab === "client-settings" && <ClientSettingsView />}
			{activeTab === "pricing-config" && <PricingConfigView />}
			{activeTab === "logging" && <LoggingView />}
			{activeTab === "governance" && <GovernanceView />}
			{activeTab === "caching" && <CachingView />}
			{activeTab === "observability" && <ObservabilityView />}
			{activeTab === "security" && <SecurityView />}
			{activeTab === "proxy" && IS_ENTERPRISE && <ProxyView />}
			{activeTab === "api-keys" && <APIKeysView />}
			{activeTab === "performance-tuning" && <PerformanceTuningView />}
			{/* Fallback to client settings for invalid tabs */}
			{activeTab && !validTabs.includes(activeTab) && <ClientSettingsView />}
		</div>
	);
}
