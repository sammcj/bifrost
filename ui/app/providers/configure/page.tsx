"use client";

import { closeConfigureDialog, useAppDispatch, useAppSelector, useGetProvidersQuery } from "@/lib/store";
import { useRouter } from "next/navigation";
import ProviderForm from "./provider-form";

export default function ConfigurePage() {
	const router = useRouter();
	const dispatch = useAppDispatch();
	const selectedProvider = useAppSelector((state) => state.provider.selectedProvider);
	const { data: providersData, refetch } = useGetProvidersQuery();

	const handleSave = () => {
		refetch();
		dispatch(closeConfigureDialog());
		router.push("/providers");
	};

	const handleCancel = () => {
		dispatch(closeConfigureDialog());
		router.push("/providers");
	};

	return (
		<div className="container mx-auto py-8">
			<ProviderForm
				provider={selectedProvider}
				allProviders={providersData?.providers || []}
				existingProviders={providersData?.providers?.map((p) => p.name) || []}
				onSave={handleSave}
				onCancel={handleCancel}
			/>
		</div>
	);
}
