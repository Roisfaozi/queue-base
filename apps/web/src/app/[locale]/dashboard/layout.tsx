import { cookies } from "next/headers";
import { DashboardShellProvider } from "./_components/dashboard-shell-context";

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
		// DEV MODE: Mock organizations to bypass backend fetch
		initialOrgs = [
			{
				id: "org-1",
				name: "Dev Organization",
				slug: "dev-org",
				status: "active",
				owner_id: "current-user",
				created_at: Date.now(),
				updated_at: Date.now(),
			},
		];
		/*
		const resp = await organizationsApi.getMyOrganizations();
		initialOrgs = resp.data?.organizations;
		*/
	} catch (error) {
		console.error("Failed to fetch initial orgs on server", error);
	}

	return (
		<DashboardShellProvider
			initialData={initialOrgs as any}
			initialSelectedOrganizationId={initialSelectedOrganizationId}
		>
			<DashboardLayoutClient>{children}</DashboardLayoutClient>
		</DashboardShellProvider>
	);
}
