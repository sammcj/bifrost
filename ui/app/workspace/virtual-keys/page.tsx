"use client";

import FullPageLoader from "@/components/fullPageLoader";
import {
	getErrorMessage,
	useGetCustomersQuery,
	useGetTeamsQuery,
	useGetVirtualKeysQuery,
	useLazyGetCoreConfigQuery,
} from "@/lib/store";
import { RbacOperation, RbacResource, useRbac } from "@enterprise/lib";
import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";
import VirtualKeysTable from "./views/virtualKeysTable";

const POLLING_INTERVAL = 5000;

export default function VirtualKeysPage() {
	const [governanceEnabled, setGovernanceEnabled] = useState<boolean | null>(null);
	const hasVirtualKeysAccess = useRbac(RbacResource.VirtualKeys, RbacOperation.View);
	const hasTeamsAccess = useRbac(RbacResource.Teams, RbacOperation.View);
	const hasCustomersAccess = useRbac(RbacResource.Customers, RbacOperation.View);
	const shownErrorsRef = useRef(new Set<string>());

	const [triggerGetConfig] = useLazyGetCoreConfigQuery();

	// Use regular queries with skip option and polling
	const {
		data: virtualKeysData,
		error: vkError,
		isLoading: vkLoading,
	} = useGetVirtualKeysQuery(undefined, {
		skip: !governanceEnabled || !hasVirtualKeysAccess,
		pollingInterval: POLLING_INTERVAL,
	});

	const {
		data: teamsData,
		error: teamsError,
		isLoading: teamsLoading,
	} = useGetTeamsQuery({}, {
		skip: !governanceEnabled || !hasTeamsAccess,
		pollingInterval: POLLING_INTERVAL,
	});

	const {
		data: customersData,
		error: customersError,
		isLoading: customersLoading,
	} = useGetCustomersQuery(undefined, {
		skip: !governanceEnabled || !hasCustomersAccess,
		pollingInterval: POLLING_INTERVAL,
	});

	const isLoading = governanceEnabled === null || (governanceEnabled && (vkLoading || teamsLoading || customersLoading));

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

	// Handle query errors - show consolidated error if all APIs fail
	useEffect(() => {
		const errorKey = `${!!vkError}-${!!teamsError}-${!!customersError}`;
		if (shownErrorsRef.current.has(errorKey)) return;
		shownErrorsRef.current.add(errorKey);
		if (vkError && teamsError && customersError) {
			// If all three APIs fail, suggest resetting bifrost
			toast.error("Failed to load governance data. Please reset Bifrost to enable governance properly.");
		} else {
			// Show individual errors if only some APIs fail
			if (vkError) {
				toast.error(`Failed to load virtual keys: ${getErrorMessage(vkError)}`);
			}
			if (teamsError) {
				toast.error(`Failed to load teams: ${getErrorMessage(teamsError)}`);
			}
			if (customersError) {
				toast.error(`Failed to load customers: ${getErrorMessage(customersError)}`);
			}
		}
	}, [vkError, teamsError, customersError]);

	if (isLoading) {
		return <FullPageLoader />;
	}

	return (
		<div className="mx-auto w-full max-w-7xl">
			<VirtualKeysTable
				virtualKeys={virtualKeysData?.virtual_keys || []}
				teams={teamsData?.teams || []}
				customers={customersData?.customers || []}
			/>
		</div>
	);
}