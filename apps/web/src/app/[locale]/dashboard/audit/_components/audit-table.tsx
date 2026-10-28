"use client";

import { useAudit } from "./audit-context";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "~/components/ui/table";
import { Badge } from "~/components/ui/badge";
import { Icon } from "~/components/shared/icon";
import { LogDetailDialog } from "~/components/dashboard/audit/log-detail-dialog";
import type { AuditLog } from "~/lib/api/audit";
import { memo } from "react";
import { EmptyState } from "~/components/shared/empty-state";

const idMonths = [
	"Jan",
	"Feb",
	"Mar",
	"Apr",
	"Mei",
	"Jun",
	"Jul",
	"Agu",
	"Sep",
	"Okt",
	"Nov",
	"Des",
];

function formatAuditTimestamp(timestamp: number): string {
	if (!timestamp) return "-";
	const date = new Date(timestamp);
	if (Number.isNaN(date.getTime())) return "-";

	const day = date.getDate();
	const month = idMonths[date.getMonth()];
	const year = date.getFullYear();
	const hours = String(date.getHours()).padStart(2, "0");
	const minutes = String(date.getMinutes()).padStart(2, "0");

	return `${day} ${month} ${year}, ${hours}.${minutes}`;
}

function normalizeAction(action: string): string {
	if (!action) return "-";
	const upper = action.toUpperCase();
	if (
		upper === "UPDATE" ||
		upper === "EDIT" ||
		upper.includes("UPDATE") ||
		upper.includes("EDIT")
	) {
		return "UBAH";
	}
	return upper;
}

export function AuditTable() {
	const {
		logs,
		isLoading,
		handleRowClick,
		selectedLog,
		isDetailOpen,
		setIsDetailOpen,
		searchTerm,
		clearSearch,
	} = useAudit();

	if (!isLoading && logs.length === 0) {
		return (
			<div className="bg-muted/5 rounded-md border border-dashed">
				{searchTerm ? (
					<EmptyState
						case="search"
						searchTerm={searchTerm}
						action={{ label: "Clear search", onClick: clearSearch }}
					/>
				) : (
					<EmptyState case="activity" />
				)}
			</div>
		);
	}

	return (
		<>
			<div className="bg-card rounded-md border">
				<Table>
					<TableHeader>
						<TableRow className="bg-muted/50">
							<TableHead className="w-[180px] uppercase text-xs font-semibold">
								Timestamp
							</TableHead>
							<TableHead className="uppercase text-xs font-semibold">
								Pelaku
							</TableHead>
							<TableHead className="uppercase text-xs font-semibold">
								Action
							</TableHead>
							<TableHead className="uppercase text-xs font-semibold">
								Entity
							</TableHead>
							<TableHead className="uppercase text-xs font-semibold">
								IP Address
							</TableHead>
							<TableHead className="w-[50px]"></TableHead>
						</TableRow>
					</TableHeader>
					<TableBody>
						{isLoading && logs.length === 0 ? (
							<TableRow>
								<TableCell colSpan={6} className="h-24 text-center">
									<div className="flex items-center justify-center gap-2">
										<Icon name="Loader" className="h-4 w-4 animate-spin" />
										Loading logs...
									</div>
								</TableCell>
							</TableRow>
						) : (
							logs.map((log) => (
								<MemoizedAuditRow
									key={log.id}
									log={log}
									onClick={() => handleRowClick(log)}
								/>
							))
						)}
					</TableBody>
				</Table>
			</div>

			<LogDetailDialog
				log={selectedLog}
				open={isDetailOpen}
				onOpenChange={setIsDetailOpen}
			/>
		</>
	);
}

const MemoizedAuditRow = memo(function AuditRow({
	log,
	onClick,
}: {
	log: AuditLog;
	onClick: () => void;
}) {
	const actor = log.actor || log.username || log.user_id || "-";
	const actionDisplay = normalizeAction(log.action);

	return (
		<TableRow
			className="hover:bg-muted/50 group cursor-pointer transition-colors"
			onClick={onClick}
		>
			<TableCell className="text-muted-foreground text-xs whitespace-nowrap">
				{formatAuditTimestamp(log.created_at)}
			</TableCell>
			<TableCell className="text-xs font-semibold text-foreground">
				{actor}
			</TableCell>
			<TableCell>
				<Badge
					variant="outline"
					className="bg-primary/10 text-primary border-primary/20 font-mono text-[10px] font-semibold uppercase px-2 py-0.5"
				>
					{actionDisplay}
				</Badge>
			</TableCell>
			<TableCell className="text-xs font-medium text-foreground">
				{log.entity}
			</TableCell>
			<TableCell className="text-muted-foreground text-xs font-mono">
				{log.ip_address || ""}
			</TableCell>
			<TableCell className="text-right pr-4">
				<Icon
					name="ChevronRight"
					className="text-muted-foreground h-4 w-4 opacity-70 group-hover:opacity-100 transition-opacity"
				/>
			</TableCell>
		</TableRow>
	);
});
