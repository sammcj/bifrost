"use client";

import { RbacOperation, RbacResource, useRbac } from "@enterprise/lib";

export default function UserGroupsLayout({ children }: { children: React.ReactNode }) {
	const hasUserGroupsAccess = useRbac(RbacResource.VirtualKeys, RbacOperation.View);
	if (!hasUserGroupsAccess) {
		return <div>You don't have permission to view virtual keys and related configurations.</div>;
	}
	return <div>{children}</div>;
}
