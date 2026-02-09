import { Button } from "@/components/ui/button";
import { useGetCoreConfigQuery } from "@/lib/store";
import { ModelProvider } from "@/lib/types/config";
import { RbacOperation, RbacResource, useRbac } from "@enterprise/lib";
import { SettingsIcon } from "lucide-react";
import { useMemo, useState } from "react";
import ProviderConfigSheet from "../dialogs/providerConfigSheet";
import ModelProviderKeysTableView from "./modelProviderKeysTableView";
import ProviderGovernanceTable from "./providerGovernanceTable";
import { keysRequired } from "./utils";

interface Props {
	provider: ModelProvider;
}

export default function ModelProviderConfig({ provider }: Props) {
	const [showConfigSheet, setShowConfigSheet] = useState(false);
	const hasGovernanceAccess = useRbac(RbacResource.Governance, RbacOperation.View);
	const { data: coreConfig } = useGetCoreConfigQuery({});
	const isGovernanceEnabled = coreConfig?.client_config?.enable_governance || false;

	const showApiKeys = useMemo(() => {
		if (provider.custom_provider_config) {
			return !(provider.custom_provider_config?.is_key_less ?? false);
		}
		return keysRequired(provider.name);
	}, [provider.name, provider.custom_provider_config?.is_key_less]);

	const editConfigButton = (
		<Button variant="outline" onClick={() => setShowConfigSheet(true)}>
			<SettingsIcon className="h-4 w-4" />
			Edit Provider Config
		</Button>
	);

	return (
		<div className="flex w-full flex-col gap-2">
			<ProviderConfigSheet show={showConfigSheet} onCancel={() => setShowConfigSheet(false)} provider={provider} />
			<ModelProviderKeysTableView provider={provider} headerActions={editConfigButton} isKeyless={!showApiKeys} />
			{hasGovernanceAccess && isGovernanceEnabled ? <ProviderGovernanceTable className="mt-4" provider={provider} /> : null}
		</div>
	);
}
