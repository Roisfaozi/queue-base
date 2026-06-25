"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import {
	Form,
	FormControl,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "~/components/ui/form";
import { settingsApi, type Setting } from "~/lib/api/qms";
import { Icon } from "~/components/shared/icon";
import { Badge } from "~/components/ui/badge";

const resolveSchema = z.object({
	key: z.string().min(1, "Key is required"),
	branch_id: z.string().optional(),
	service_id: z.string().optional(),
	counter_id: z.string().optional(),
});

type ResolveFormValues = z.infer<typeof resolveSchema>;

export function ResolvePanel() {
	const [isLoading, setIsLoading] = useState(false);
	const [result, setResult] = useState<Setting | null>(null);
	const [error, setError] = useState<string | null>(null);

	const form = useForm<ResolveFormValues>({
		resolver: zodResolver(resolveSchema),
		defaultValues: {
			key: "",
			branch_id: "",
			service_id: "",
			counter_id: "",
		},
	});

	async function onSubmit(data: ResolveFormValues) {
		setIsLoading(true);
		setError(null);
		setResult(null);
		try {
			const resp = await settingsApi.resolve({
				key: data.key,
				branch_id: data.branch_id || undefined,
				service_id: data.service_id || undefined,
				counter_id: data.counter_id || undefined,
			});
			setResult(resp.data);
		} catch (err: any) {
			setError(err.message || "Failed to resolve setting");
		} finally {
			setIsLoading(false);
		}
	}

	return (
		<div className="rounded-lg border bg-card text-card-foreground shadow-sm">
			<div className="flex flex-col space-y-1.5 p-6">
				<h3 className="font-semibold leading-none tracking-tight">
					Test Resolution
				</h3>
				<p className="text-sm text-muted-foreground">
					Simulate how the backend resolves a setting key based on context.
				</p>
			</div>
			<div className="p-6 pt-0">
				<Form {...form}>
					<form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
						<div className="grid grid-cols-2 gap-4">
							<FormField
								control={form.control}
								name="key"
								render={({ field }) => (
									<FormItem>
										<FormLabel>Setting Key</FormLabel>
										<FormControl>
											<Input placeholder="e.g. ticket_prefix" {...field} />
										</FormControl>
										<FormMessage />
									</FormItem>
								)}
							/>
							<FormField
								control={form.control}
								name="branch_id"
								render={({ field }) => (
									<FormItem>
										<FormLabel>Branch ID (Optional)</FormLabel>
										<FormControl>
											<Input placeholder="uuid..." {...field} />
										</FormControl>
										<FormMessage />
									</FormItem>
								)}
							/>
							<FormField
								control={form.control}
								name="service_id"
								render={({ field }) => (
									<FormItem>
										<FormLabel>Service ID (Optional)</FormLabel>
										<FormControl>
											<Input placeholder="uuid..." {...field} />
										</FormControl>
										<FormMessage />
									</FormItem>
								)}
							/>
							<FormField
								control={form.control}
								name="counter_id"
								render={({ field }) => (
									<FormItem>
										<FormLabel>Counter ID (Optional)</FormLabel>
										<FormControl>
											<Input placeholder="uuid..." {...field} />
										</FormControl>
										<FormMessage />
									</FormItem>
								)}
							/>
						</div>
						<Button type="submit" disabled={isLoading}>
							{isLoading && (
								<Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
							)}
							Resolve Value
						</Button>
					</form>
				</Form>

				{error && (
					<div className="mt-6 rounded-md bg-destructive/15 p-4 text-sm text-destructive border border-destructive/20">
						{error}
					</div>
				)}

				{result && (
					<div className="mt-6 rounded-md border p-4">
						<h4 className="mb-2 text-sm font-semibold">Resolution Result</h4>
						<div className="grid grid-cols-2 gap-2 text-sm">
							<div className="text-muted-foreground">Resolved Value:</div>
							<div className="font-mono font-medium">{result.value}</div>
							<div className="text-muted-foreground">Source Scope:</div>
							<div>
								<Badge variant="outline">{result.scope_type}</Badge>
							</div>
							<div className="text-muted-foreground">Setting ID:</div>
							<div className="font-mono text-xs">{result.id}</div>
						</div>
					</div>
				)}
			</div>
		</div>
	);
}
