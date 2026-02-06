"use client";

import FullPageLoader from "@/components/fullPageLoader";
import { getErrorMessage, useGetModelConfigsQuery, useLazyGetCoreConfigQuery } from "@/lib/store";
import { RbacOperation, RbacResource, useRbac } from "@enterprise/lib";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import ModelLimitsTable from "./modelLimitsTable";

export default function ModelLimitsView() {
	const [governanceEnabled, setGovernanceEnabled] = useState<boolean | null>(null);
	const hasGovernanceAccess = useRbac(RbacResource.Governance, RbacOperation.View);

	const [triggerGetConfig] = useLazyGetCoreConfigQuery();

	// Use regular query with skip and polling
	const {
		data: modelConfigsData,
		error: modelConfigsError,
		isLoading: modelConfigsLoading,
	} = useGetModelConfigsQuery(undefined, {
		skip: !governanceEnabled || !hasGovernanceAccess,
		pollingInterval: 5000,
	});

	const isLoading = modelConfigsLoading || governanceEnabled === null;

	useEffect(() => {
		triggerGetConfig({ fromDB: true })
			.then((res) => {
				if (res.data?.client_config?.enable_governance) {
					setGovernanceEnabled(true);
				} else {
					setGovernanceEnabled(false);
					toast.error("Governance is not enabled. Please enable it in the config.");
				}
			})
			.catch((err) => {
				console.error("Failed to fetch config:", err);
				setGovernanceEnabled(false);
				toast.error(getErrorMessage(err) || "Failed to load configuration");
			});
	}, [triggerGetConfig]);

	// Handle query errors
	useEffect(() => {
		if (modelConfigsError) {
			toast.error(`Failed to load model configs: ${getErrorMessage(modelConfigsError)}`);
		}
	}, [modelConfigsError]);

	if (isLoading) {
		return <FullPageLoader />;
	}

	return (
		<div className="mx-auto w-full max-w-7xl">
			<ModelLimitsTable modelConfigs={modelConfigsData?.model_configs || []} />
		</div>
	);
}
