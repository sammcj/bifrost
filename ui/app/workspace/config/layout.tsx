"use client";

import { RbacOperation, RbacResource, useRbac } from "@enterprise/lib";

export default function ConfigLayout({ children }: { children: React.ReactNode }) {
	const hasConfigAccess = useRbac(RbacResource.Settings, RbacOperation.View);
	if (!hasConfigAccess) {
		return <div>You don't have permission to view config</div>;
	}
	return <div>{children}</div>;
}
