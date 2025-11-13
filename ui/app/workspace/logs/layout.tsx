"use client";

import { RbacOperation, RbacResource, useRbac } from "@enterprise/lib";

export default function LogsLayout({ children }: { children: React.ReactNode }) {
	const hasViewLogsAccess = useRbac(RbacResource.Logs, RbacOperation.View);
	if (!hasViewLogsAccess) {
		return <div>You don't have permission to view logs</div>;
	}
	return <div>{children}</div>;
}
