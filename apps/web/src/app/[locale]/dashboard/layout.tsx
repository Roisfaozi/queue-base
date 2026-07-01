import { cookies } from "next/headers";
import { DashboardShellProvider } from "./_components/dashboard-shell-context";
import { ApiError } from "~/lib/api/client";
import { organizationsApi } from "~/lib/api/organizations";

// Better way: import usePresence in a dedicated client component or just here if we make this function client
// Let's create a Client wrapper for the layout logic
import { DashboardLayoutClient } from "./layout-client";

export default async function DashboardLayout({
	children,
}: {
	children: React.ReactNode;
}) {
	const cookieStore = await cookies();
	const initialSelectedOrganizationId =
		cookieStore.get("organization_id")?.value || null;

	// 1. Fetch organizations on Server (Critical for Navigation/Switcher)
	let initialOrgs = undefined;
	try {
		const resp = await organizationsApi.getMyOrganizations();
		initialOrgs = resp.data?.organizations;
	} catch (error) {
		if (
			(error instanceof ApiError && error.status === 401) ||
			(error instanceof Error &&
				(error.message === "Session expired" ||
					error.message === "Unauthenticated"))
		) {
			console.warn("Initial organizations unavailable: unauthenticated");
		} else {
			console.error("Failed to fetch initial orgs on server", error);
		}
	}

	return (
		<DashboardShellProvider
			initialData={initialOrgs}
			initialSelectedOrganizationId={initialSelectedOrganizationId}
		>
			<DashboardLayoutClient>{children}</DashboardLayoutClient>
		</DashboardShellProvider>
	);
}
