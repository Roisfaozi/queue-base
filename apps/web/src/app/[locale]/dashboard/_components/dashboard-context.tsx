"use client";

import {
	createContext,
	useContext,
	useState,
	useEffect,
	useCallback,
	type ReactNode,
} from "react";
import { statsApi, type SystemInsights } from "~/lib/api/stats";
import { auditApi, type AuditLog } from "~/lib/api/audit";
import { toast } from "sonner";
import { useDashboardShell } from "./dashboard-shell-context";

interface DashboardContextType {
	stats: {
		users: number;
		roles: number;
		auditLogs: number;
	};
	insights: SystemInsights | null;
	recentLogs: AuditLog[];
	isLoading: boolean;
	hasOrganization: boolean;
	refresh: () => Promise<void>;
}

const DashboardContext = createContext<DashboardContextType | undefined>(
	undefined,
);

export function DashboardProvider({ children }: { children: ReactNode }) {
	const { currentOrganization, isLoading: isShellLoading } =
		useDashboardShell();
	const [stats, setStats] = useState({
		users: 0,
		roles: 0,
		auditLogs: 0,
	});
	const [insights, setInsights] = useState<SystemInsights | null>(null);
	const [recentLogs, setRecentLogs] = useState<AuditLog[]>([]);
	const [isLoading, setIsLoading] = useState(true);

	const fetchData = useCallback(async () => {
		if (isShellLoading) {
			return;
		}

		if (!currentOrganization) {
			setStats({ users: 0, roles: 0, auditLogs: 0 });
			setInsights(null);
			setRecentLogs([]);
			setIsLoading(false);
			return;
		}

		setIsLoading(true);
		try {
			const [summaryResp, insightsResp, recentLogsResp] = await Promise.all([
				statsApi.getSummary(),
				statsApi.getInsights(),
				auditApi.search({
					page: 1,
					page_size: 5,
					sort: [{ colId: "created_at", sort: "desc" }],
				}),
			]);

			if (summaryResp) {
				setStats({
					users: summaryResp.total_users,
					roles: summaryResp.total_roles,
					auditLogs: summaryResp.total_audit_logs,
				});
			}

			if (insightsResp) {
				setInsights(insightsResp);
			}

			if (recentLogsResp.data) {
				setRecentLogs(recentLogsResp.data);
			}
		} catch (error) {
			console.error("Dashboard fetch error:", error);
			toast.error("Failed to load dashboard data");
		} finally {
			setIsLoading(false);
		}
	}, [currentOrganization, isShellLoading]);

	useEffect(() => {
		if (isShellLoading) {
			setIsLoading(true);
			return;
		}

		fetchData();
	}, [fetchData, isShellLoading]);

	return (
		<DashboardContext.Provider
			value={{
				stats,
				insights,
				recentLogs,
				isLoading,
				hasOrganization: !!currentOrganization,
				refresh: fetchData,
			}}
		>
			{children}
		</DashboardContext.Provider>
	);
}

export function useDashboard() {
	const context = useContext(DashboardContext);
	if (context === undefined) {
		throw new Error("useDashboard must be used within a DashboardProvider");
	}
	return context;
}
