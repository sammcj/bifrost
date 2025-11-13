"use client";

import { RbacOperation, RbacResource, useRbac } from "@enterprise/lib";

export default function VirtualKeysLayout({ children }: { children: React.ReactNode }) {
	const hasVirtualKeysAccess = useRbac(RbacResource.VirtualKeys, RbacOperation.View);
	if (!hasVirtualKeysAccess) {
		return <div>You don't have permission to view virtual keys and related configurations.</div>;
	}
	return <div>{children}</div>;
}
