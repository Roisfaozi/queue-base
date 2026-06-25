"use client";

import {
	createContext,
	useContext,
	useCallback,
	useEffect,
	type ReactNode,
} from "react";
import { type Organization, organizationsApi } from "~/lib/api/organizations";
import { useOrganizationStore } from "~/stores/use-organization-store";
import { toast } from "sonner";
import useSWR from "swr";

interface DashboardShellContextType {
	organizations: Organization[];
	currentOrganization: Organization | null;
	isLoading: boolean;
	setOrganization: (org: Organization) => void;
	refreshOrganizations: () => Promise<void>;
}

const DashboardShellContext = createContext<
	DashboardShellContextType | undefined
>(undefined);

export function DashboardShellProvider({
	children,
	initialData,
	initialSelectedOrganizationId,
}: {
	children: ReactNode;
	initialData?: Organization[];
	initialSelectedOrganizationId?: string | null;
}) {
	const { currentOrganization, setCurrentOrganization } =
		useOrganizationStore();

	const syncSelectedOrganization = useCallback(
		async (organization: Organization | null) => {
			const method = organization ? "PATCH" : "DELETE";
			const body = organization
				? JSON.stringify({
						organizationId: organization.id,
						organizationSlug: organization.slug,
					})
				: undefined;

			await fetch("/api/organizations/current", {
				method,
				headers: body ? { "Content-Type": "application/json" } : undefined,
				body,
				credentials: "include",
			});
		},
		[],
	);

	const {
		data: organizations = [],
		isLoading,
		mutate,
	} = useSWR(
		"/api/v1/organizations/me",
		() =>
			organizationsApi
				.getMyOrganizations()
				.then((res) => res.data?.organizations || []),
		{
			fallbackData: initialData,
			keepPreviousData: true,
		},
	);

	useEffect(() => {
		if (isLoading) return;

		if (organizations.length === 0) {
			if (currentOrganization) {
				setCurrentOrganization(null);
				void syncSelectedOrganization(null);
			}
			return;
		}

		const selectedFromStore = currentOrganization
			? organizations.find((org) => org.id === currentOrganization.id) || null
			: null;

		const selectedFromCookie = initialSelectedOrganizationId
			? organizations.find((org) => org.id === initialSelectedOrganizationId) ||
				null
			: null;

		const nextOrganization =
			selectedFromStore || selectedFromCookie || organizations[0] || null;

		if (!nextOrganization) {
			return;
		}

		if (currentOrganization?.id !== nextOrganization.id) {
			setCurrentOrganization(nextOrganization);
		}

		if (initialSelectedOrganizationId !== nextOrganization.id) {
			void syncSelectedOrganization(nextOrganization);
		}
	}, [
		currentOrganization,
		initialSelectedOrganizationId,
		isLoading,
		organizations,
		setCurrentOrganization,
		syncSelectedOrganization,
	]);

	const fetchOrgs = useCallback(async () => {
		await mutate();
	}, [mutate]);

	const setOrganization = useCallback(
		(org: Organization) => {
			setCurrentOrganization(org);
			void syncSelectedOrganization(org);
			toast.success(`Switched to ${org.name}`);
		},
		[setCurrentOrganization, syncSelectedOrganization],
	);

	return (
		<DashboardShellContext.Provider
			value={{
				organizations,
				currentOrganization,
				isLoading,
				setOrganization,
				refreshOrganizations: fetchOrgs,
			}}
		>
			{children}
		</DashboardShellContext.Provider>
	);
}

export function useDashboardShell() {
	const context = useContext(DashboardShellContext);
	if (context === undefined) {
		throw new Error(
			"useDashboardShell must be used within a DashboardShellProvider",
		);
	}
	return context;
}
