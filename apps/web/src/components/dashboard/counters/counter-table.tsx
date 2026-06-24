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
import type { Counter } from "~/lib/api/qms";

interface CounterTableProps {
	counters: Counter[];
	isLoading: boolean;
	error: any;
	canUpdate: boolean;
	canDelete: boolean;
	onEdit: (counter: Counter) => void;
	onDelete: (counter: Counter) => void;
	onCreateCounter?: () => void;
}

export function CounterTable({
	counters,
	isLoading,
	error,
	canUpdate,
	canDelete,
	onEdit,
	onDelete,
	onCreateCounter,
}: CounterTableProps) {
	if (!isLoading && !error && counters.length === 0) {
		return (
			<div className="bg-muted/5 rounded-md border border-dashed">
				<EmptyState
					case="generic"
					title="No counters found"
					description="Create your first counter for a branch."
					action={
						onCreateCounter
							? {
									label: "Add Counter",
									onClick: onCreateCounter,
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
						<TableHead>Code</TableHead>
						<TableHead>Name</TableHead>
						<TableHead>Branch ID</TableHead>
						<TableHead>Status</TableHead>
						<TableHead className="text-right">Created</TableHead>
						{(canUpdate || canDelete) && (
							<TableHead className="w-[50px]"></TableHead>
						)}
					</TableRow>
				</TableHeader>
				<TableBody>
					{isLoading ? (
						<TableRow>
							<TableCell colSpan={6} className="h-24 text-center">
								<div className="flex items-center justify-center gap-2">
									<Icon name="Loader" className="h-4 w-4 animate-spin" />
									Loading counters...
								</div>
							</TableCell>
						</TableRow>
					) : error ? (
						<TableRow>
							<TableCell colSpan={6} className="h-24 text-center">
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
						counters.map((counter) => (
							<MemoizedCounterTableRow
								key={counter.id}
								counter={counter}
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

const MemoizedCounterTableRow = memo(function CounterTableRow({
	counter,
	canUpdate,
	canDelete,
	onEdit,
	onDelete,
}: {
	counter: Counter;
	canUpdate: boolean;
	canDelete: boolean;
	onEdit: (counter: Counter) => void;
	onDelete: (counter: Counter) => void;
}) {
	return (
		<TableRow>
			<TableCell className="font-medium">
				<Badge variant="outline" className="font-mono">
					{counter.code}
				</Badge>
			</TableCell>
			<TableCell className="font-medium">{counter.name}</TableCell>
			<TableCell className="text-xs text-muted-foreground font-mono">
				{counter.branch_id}
			</TableCell>
			<TableCell>
				<Badge
					variant={counter.status === "active" ? "default" : "secondary"}
					className={
						counter.status === "active"
							? "bg-emerald-500 hover:bg-emerald-600"
							: ""
					}
				>
					{counter.status || "Unknown"}
				</Badge>
			</TableCell>
			<TableCell className="text-right text-muted-foreground">
				{format(new Date(counter.created_at * 1000), "MMM dd, yyyy")}
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
								<DropdownMenuItem onClick={() => onEdit(counter)}>
									<Icon name="Pencil" className="mr-2 h-4 w-4" />
									Edit
								</DropdownMenuItem>
							)}
							{canDelete && (
								<DropdownMenuItem
									onClick={() => onDelete(counter)}
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
