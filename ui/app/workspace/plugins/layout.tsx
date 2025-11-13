"use client";

import { RbacOperation, RbacResource, useRbac } from "@enterprise/lib";

export default function PluginsLayout({ children }: { children: React.ReactNode }) {
	const hasPluginsAccess = useRbac(RbacResource.Plugins, RbacOperation.View);
	if (!hasPluginsAccess) {
		return <div>You don't have permission to view plugins</div>;
	}
	return <div>{children}</div>;
}