"use client";

import { RbacOperation, RbacResource, useRbac } from "@enterprise/lib";

export default function ObservabilityLayout({ children }: { children: React.ReactNode }) {
	const hasObservabilityAccess = useRbac(RbacResource.Observability, RbacOperation.View);
	if (!hasObservabilityAccess) {
		return <div>You don't have permission to view observability settings.</div>;
	}
	return <div>{children}</div>;
}
