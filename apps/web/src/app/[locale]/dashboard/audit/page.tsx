"use client";

import { Icon } from "~/components/shared/icon";
import { useDashboardShell } from "../_components/dashboard-shell-context";
import { AuditProvider } from "./_components/audit-context";
import { AuditPagination } from "./_components/audit-pagination";
import { AuditTable } from "./_components/audit-table";
import { AuditToolbar } from "./_components/audit-toolbar";

export default function AuditPage() {
	const { currentOrganization, isLoading } = useDashboardShell();

	if (isLoading) {
		return (
			<div className="flex h-[400px] items-center justify-center rounded-lg border-2 border-dashed">
				<p className="text-muted-foreground">Loading organization context...</p>
			</div>
		);
	}

	if (!currentOrganization) {
		return (
			<div className="flex h-[400px] items-center justify-center rounded-lg border-2 border-dashed">
				<div className="text-center">
					<Icon
						name="Building2"
						className="text-muted-foreground/50 mx-auto h-8 w-8"
					/>
					<p className="text-muted-foreground mt-2">
						Please select an organization first.
					</p>
				</div>
			</div>
		);
	}

	return (
		<AuditProvider>
			<div className="space-y-4">
				<AuditToolbar />
				<AuditTable />
				<AuditPagination />
			</div>
		</AuditProvider>
	);
}
