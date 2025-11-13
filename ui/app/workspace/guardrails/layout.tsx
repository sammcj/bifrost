"use client";

import { RbacOperation, RbacResource, useRbac } from "@enterprise/lib";

export default function GuardrailsLayout({ children }: { children: React.ReactNode }) {
	const hasGuardrailsAccess = useRbac(RbacResource.GuardrailsConfig, RbacOperation.View);
	if (!hasGuardrailsAccess) {
		return <div>You don't have permission to view guardrails configuration</div>;
	}
	return <div>{children}</div>;
}
