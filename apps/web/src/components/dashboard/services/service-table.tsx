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
import type { Service } from "~/lib/api/qms";

interface ServiceTableProps {
	services: Service[];
	isLoading: boolean;
	error: any;
	canUpdate: boolean;
	canDelete: boolean;
	onEdit: (service: Service) => void;
	onDelete: (service: Service) => void;
	onCreateService?: () => void;
}

export function ServiceTable({
	services,
	isLoading,
	error,
	canUpdate,
	canDelete,
	onEdit,
	onDelete,
	onCreateService,
}: ServiceTableProps) {
	if (!isLoading && !error && services.length === 0) {
		return (
			<div className="bg-muted/5 rounded-md border border-dashed">
				<EmptyState
					case="generic"
					title="No services found"
					description="Get started by creating your first queue service."
					action={
						onCreateService
							? {
									label: "Add Service",
									onClick: onCreateService,
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
						<TableHead>Pharmacy Logic</TableHead>
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
									Loading services...
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
						services.map((service) => (
							<MemoizedServiceTableRow
								key={service.id}
								service={service}
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

const MemoizedServiceTableRow = memo(function ServiceTableRow({
	service,
	canUpdate,
	canDelete,
	onEdit,
	onDelete,
}: {
	service: Service;
	canUpdate: boolean;
	canDelete: boolean;
	onEdit: (service: Service) => void;
	onDelete: (service: Service) => void;
}) {
	return (
		<TableRow>
			<TableCell className="font-medium">
				<Badge variant="outline" className="font-mono">
					{service.code}
				</Badge>
			</TableCell>
			<TableCell className="font-medium">{service.name}</TableCell>
			<TableCell>
				<div className="flex flex-col gap-1 text-xs text-muted-foreground">
					{service.is_pharmacy && (
						<span className="flex items-center gap-1">
							<Icon name="Pill" className="h-3 w-3" /> Pharmacy Flow
						</span>
					)}
					{service.is_pharmacy_reception && (
						<span className="flex items-center gap-1">
							<Icon name="Inbox" className="h-3 w-3" /> Rx Reception
						</span>
					)}
					{!service.is_pharmacy && !service.is_pharmacy_reception && "-"}
				</div>
			</TableCell>
			<TableCell>
				<Badge
					variant={service.status === "active" ? "default" : "secondary"}
					className={
						service.status === "active"
							? "bg-emerald-500 hover:bg-emerald-600"
							: ""
					}
				>
					{service.status || "Unknown"}
				</Badge>
			</TableCell>
			<TableCell className="text-right text-muted-foreground">
				{format(new Date(service.created_at * 1000), "MMM dd, yyyy")}
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
								<DropdownMenuItem onClick={() => onEdit(service)}>
									<Icon name="Pencil" className="mr-2 h-4 w-4" />
									Edit
								</DropdownMenuItem>
							)}
							{canDelete && (
								<DropdownMenuItem
									onClick={() => onDelete(service)}
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
