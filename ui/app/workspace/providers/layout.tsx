"use client";

import { RbacOperation, RbacResource, useRbac } from "@enterprise/lib";

export default function ProvidersLayout({ children }: { children: React.ReactNode }) {
	const hasProvidersAccess = useRbac(RbacResource.ModelProvider, RbacOperation.View);
	if (!hasProvidersAccess) {
		return <div>You don't have permission to view model providers</div>;
	}
	return <div>{children}</div>;
}