"use client";

import { memo } from "react";
import { format } from "date-fns";
import { EmptyState } from "~/components/shared/empty-state";
import { Icon } from "~/components/shared/icon";
import { Badge } from "~/components/ui/badge";
import { Button } from "~/components/ui/button";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuTrigger,
} from "~/components/ui/dropdown-menu";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "~/components/ui/table";
import type { Setting } from "~/lib/api/qms";

interface SettingsTableProps {
	settings: Setting[];
	isLoading: boolean;
	error: any;
	canUpdate: boolean;
	canDelete: boolean;
	onEdit: (setting: Setting) => void;
	onDelete: (setting: Setting) => void;
	onCreateSetting?: () => void;
}

const SCOPE_COLORS: Record<string, string> = {
	tenant: "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200",
	branch:
		"bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200",
	service: "bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-200",
	counter:
		"bg-emerald-100 text-emerald-800 dark:bg-emerald-900 dark:text-emerald-200",
};

export function SettingsTable({
	settings,
	isLoading,
	error,
	canUpdate,
	canDelete,
	onEdit,
	onDelete,
	onCreateSetting,
}: SettingsTableProps) {
	if (!isLoading && !error && settings.length === 0) {
		return (
			<div className="bg-muted/5 rounded-md border border-dashed">
				<EmptyState
					case="generic"
					title="No settings found"
					description="Add your first queue setting override."
					action={
						onCreateSetting
							? {
									label: "Add Setting",
									onClick: onCreateSetting,
									icon: "Plus",
								}
							: undefined
					}
				/>
			</div>
		);
	}

	return (
		<div className="rounded-md border">
			<Table>
				<TableHeader>
					<TableRow>
						<TableHead>Key</TableHead>
						<TableHead>Value</TableHead>
						<TableHead>Scope</TableHead>
						<TableHead>Scope ID</TableHead>
						<TableHead>Status</TableHead>
						<TableHead className="text-right">Updated</TableHead>
						{(canUpdate || canDelete) && (
							<TableHead className="w-[50px]"></TableHead>
						)}
					</TableRow>
				</TableHeader>
				<TableBody>
					{isLoading ? (
						<TableRow>
							<TableCell colSpan={7} className="h-24 text-center">
								<div className="flex items-center justify-center gap-2">
									<Icon name="Loader" className="h-4 w-4 animate-spin" />
									Loading settings...
								</div>
							</TableCell>
						</TableRow>
					) : error ? (
						<TableRow>
							<TableCell colSpan={7} className="h-24 text-center">
								<EmptyState
									case="error"
									action={{
										label: "Retry",
										onClick: () => window.location.reload(),
									}}
								/>
							</TableCell>
						</TableRow>
					) : (
						settings.map((setting) => (
							<MemoizedSettingsTableRow
								key={setting.id}
								setting={setting}
								canUpdate={canUpdate}
								canDelete={canDelete}
								onEdit={onEdit}
								onDelete={onDelete}
							/>
						))
					)}
				</TableBody>
			</Table>
		</div>
	);
}

const MemoizedSettingsTableRow = memo(function SettingsTableRow({
	setting,
	canUpdate,
	canDelete,
	onEdit,
	onDelete,
}: {
	setting: Setting;
	canUpdate: boolean;
	canDelete: boolean;
	onEdit: (setting: Setting) => void;
	onDelete: (setting: Setting) => void;
}) {
	return (
		<TableRow>
			<TableCell className="font-mono text-xs font-medium">
				{setting.key}
			</TableCell>
			<TableCell className="max-w-[200px] truncate font-medium">
				{setting.value}
			</TableCell>
			<TableCell>
				<Badge
					variant="outline"
					className={SCOPE_COLORS[setting.scope_type] || ""}
				>
					{setting.scope_type}
				</Badge>
			</TableCell>
			<TableCell className="text-xs text-muted-foreground font-mono">
				{setting.scope_id ? setting.scope_id.substring(0, 12) + "..." : "-"}
			</TableCell>
			<TableCell>
				<Badge
					variant={setting.is_active ? "default" : "secondary"}
					className={
						setting.is_active ? "bg-emerald-500 hover:bg-emerald-600" : ""
					}
				>
					{setting.is_active ? "Active" : "Inactive"}
				</Badge>
			</TableCell>
			<TableCell className="text-right text-muted-foreground">
				{format(new Date(setting.updated_at * 1000), "MMM dd, yyyy")}
			</TableCell>
			{(canUpdate || canDelete) && (
				<TableCell>
					<DropdownMenu>
						<DropdownMenuTrigger asChild>
							<Button variant="ghost" size="icon" className="h-8 w-8">
								<Icon name="Ellipsis" className="h-4 w-4" />
								<span className="sr-only">Open menu</span>
							</Button>
						</DropdownMenuTrigger>
						<DropdownMenuContent align="end">
							{canUpdate && (
								<DropdownMenuItem onClick={() => onEdit(setting)}>
									<Icon name="Pencil" className="mr-2 h-4 w-4" />
									Edit
								</DropdownMenuItem>
							)}
							{canDelete && (
								<DropdownMenuItem
									onClick={() => onDelete(setting)}
									className="text-destructive"
								>
									<Icon name="Trash" className="mr-2 h-4 w-4" />
									Delete
								</DropdownMenuItem>
							)}
						</DropdownMenuContent>
					</DropdownMenu>
				</TableCell>
			)}
		</TableRow>
	);
});
