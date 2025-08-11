"use client";

import ProvidersList from "@/app/providers/views/providers-list";
import FullPageLoader from "@/components/full-page-loader";
import { useToast } from "@/hooks/use-toast";
import { getErrorMessage, useGetProvidersQuery } from "@/lib/store";
import { useEffect } from "react";

export default function Providers() {
	const { data, error, isLoading, refetch } = useGetProvidersQuery();

	const { toast } = useToast();

	useEffect(() => {
		if (error) {
			toast({
				title: "Error",
				description: getErrorMessage(error),
				variant: "destructive",
			});
		}
	}, [error, toast]);

	if (isLoading) {
		return <FullPageLoader />;
	}

	return (
		<div>
			<ProvidersList providers={data?.providers || []} onRefresh={() => refetch()} />
		</div>
	);
}
