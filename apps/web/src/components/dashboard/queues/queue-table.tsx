"use client";

import { format } from "date-fns";
import { memo } from "react";
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
import type { Queue } from "~/lib/api/qms";

const STATUS_COLORS: Record<string, string> = {
	waiting: "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200",
	calling:
		"bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200",
	serving:
		"bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200",
	skipped: "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200",
	completed:
		"bg-emerald-100 text-emerald-800 dark:bg-emerald-900 dark:text-emerald-200",
	canceled: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200",
};

interface QueueTableProps {
	queues: Queue[];
	isLoading: boolean;
	error: any;
	canTransition: boolean;
	canForward: boolean;
	onTransition: (
		queue: Queue,
		action: "call" | "serve" | "complete" | "skip" | "cancel",
	) => void;
	onForward: (queue: Queue) => void;
	onView: (queue: Queue) => void;
	onRegister?: () => void;
}

export function QueueTable({
	queues,
	isLoading,
	error,
	canTransition,
	canForward,
	onTransition,
	onForward,
	onView,
	onRegister,
}: QueueTableProps) {
	if (!isLoading && !error && queues.length === 0) {
		return (
			<div className="bg-muted/5 rounded-md border border-dashed">
				<EmptyState
					case="generic"
					title="No queues found"
					description="Register a new patient queue entry."
					action={
						onRegister
							? { label: "Register Queue", onClick: onRegister, icon: "Plus" }
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
						<TableHead>Ticket No</TableHead>
						<TableHead>Queue No</TableHead>
						<TableHead>Patient</TableHead>
						<TableHead>Status</TableHead>
						<TableHead>Date</TableHead>
						<TableHead className="text-right">Updated</TableHead>
						<TableHead className="w-[50px]"></TableHead>
					</TableRow>
				</TableHeader>
				<TableBody>
					{isLoading ? (
						<TableRow>
							<TableCell colSpan={7} className="h-24 text-center">
								<div className="flex items-center justify-center gap-2">
									<Icon name="Loader" className="h-4 w-4 animate-spin" />
									Loading queues...
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
						queues.map((queue) => (
							<MemoizedQueueRow
								key={queue.id}
								queue={queue}
								canTransition={canTransition}
								canForward={canForward}
								onTransition={onTransition}
								onForward={onForward}
								onView={onView}
							/>
						))
					)}
				</TableBody>
			</Table>
		</div>
	);
}

const MemoizedQueueRow = memo(function QueueRow({
	queue,
	canTransition,
	canForward,
	onTransition,
	onForward,
	onView,
}: {
	queue: Queue;
	canTransition: boolean;
	canForward: boolean;
	onTransition: (
		queue: Queue,
		action: "call" | "serve" | "complete" | "skip" | "cancel",
	) => void;
	onForward: (queue: Queue) => void;
	onView: (queue: Queue) => void;
}) {
	const actions = getAvailableActions(queue.status);

	return (
		<TableRow className="cursor-pointer" onClick={() => onView(queue)}>
			<TableCell className="font-mono text-xs font-semibold">
				{queue.ticket_no}
			</TableCell>
			<TableCell className="font-mono text-xl font-bold">
				{queue.queue_no}
			</TableCell>
			<TableCell>
				<div className="text-sm font-medium">{queue.patient_name || "-"}</div>
				{queue.patient_id && (
					<div className="text-xs text-muted-foreground font-mono">
						{queue.patient_id.substring(0, 8)}...
					</div>
				)}
			</TableCell>
			<TableCell>
				<Badge variant="outline" className={STATUS_COLORS[queue.status] || ""}>
					{queue.status}
				</Badge>
			</TableCell>
			<TableCell className="text-sm text-muted-foreground">
				{queue.queue_date}
			</TableCell>
			<TableCell className="text-right text-muted-foreground text-sm">
				{format(new Date(queue.updated_at * 1000), "HH:mm")}
			</TableCell>
			<TableCell onClick={(e) => e.stopPropagation()}>
				<DropdownMenu>
					<DropdownMenuTrigger asChild>
						<Button variant="ghost" size="icon" className="h-8 w-8">
							<Icon name="Ellipsis" className="h-4 w-4" />
							<span className="sr-only">Actions</span>
						</Button>
					</DropdownMenuTrigger>
					<DropdownMenuContent align="end">
						<DropdownMenuItem onClick={() => onView(queue)}>
							<Icon name="Eye" className="mr-2 h-4 w-4" />
							View Details
						</DropdownMenuItem>
						{canTransition &&
							actions.map((action) => (
								<DropdownMenuItem
									key={action}
									onClick={() => onTransition(queue, action)}
								>
									<Icon
										name={
											action === "call"
												? "Megaphone"
												: action === "serve"
													? "Play"
													: action === "complete"
														? "Check"
														: action === "skip"
															? "SkipForward"
															: "X"
										}
										className="mr-2 h-4 w-4"
									/>
									{action.charAt(0).toUpperCase() + action.slice(1)}
								</DropdownMenuItem>
							))}
						{canForward && (
							<DropdownMenuItem onClick={() => onForward(queue)}>
								<Icon name="Forward" className="mr-2 h-4 w-4" />
								Forward
							</DropdownMenuItem>
						)}
					</DropdownMenuContent>
				</DropdownMenu>
			</TableCell>
		</TableRow>
	);
});

function getAvailableActions(
	status: string,
): ("call" | "serve" | "complete" | "skip" | "cancel")[] {
	switch (status) {
		case "waiting":
			return ["call", "skip", "cancel"];
		case "calling":
			return ["serve", "skip", "cancel"];
		case "serving":
			return ["complete", "cancel"];
		case "skipped":
			return ["call", "cancel"];
		default:
			return [];
	}
}
