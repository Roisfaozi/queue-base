"use client";

import { useEffect, useState, useCallback } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "~/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from "~/components/ui/dialog";
import {
	Form,
	FormControl,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "~/components/ui/form";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "~/components/ui/select";
import { toast } from "sonner";
import { Icon } from "~/components/shared/icon";
import {
	queuesApi,
	servicesApi,
	type Service,
	type Queue,
} from "~/lib/api/qms";

const forwardSchema = z.object({
	destination_service_id: z.string().min(1, "Destination service is required."),
});

type ForwardFormValues = z.infer<typeof forwardSchema>;

interface QueueForwardDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	queue?: Queue | null;
	onSuccess: () => void;
}

export function QueueForwardDialog({
	open,
	onOpenChange,
	queue,
	onSuccess,
}: QueueForwardDialogProps) {
	const [isLoading, setIsLoading] = useState(false);
	const [services, setServices] = useState<Service[]>([]);

	const form = useForm<ForwardFormValues>({
		resolver: zodResolver(forwardSchema),
		defaultValues: { destination_service_id: "" },
	});

	const fetchServices = useCallback(async () => {
		try {
			const resp = await servicesApi.getAll();
			setServices(resp.data?.filter((s) => s.status === "active") || []);
		} catch {
			// silently ignore
		}
	}, []);

	useEffect(() => {
		if (open) {
			form.reset({ destination_service_id: "" });
			fetchServices();
		}
	}, [open, form, fetchServices]);

	async function onSubmit(data: ForwardFormValues) {
		if (!queue) return;
		setIsLoading(true);
		try {
			await queuesApi.forward(queue.id, {
				destination_service_id: data.destination_service_id,
			});
			toast.success(`Queue ${queue.ticket_no} forwarded`);
			onSuccess();
			onOpenChange(false);
		} catch (error: any) {
			toast.error(error.message || "Failed to forward queue");
		} finally {
			setIsLoading(false);
		}
	}

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="sm:max-w-[425px]">
				<DialogHeader>
					<DialogTitle>Forward Queue</DialogTitle>
				</DialogHeader>
				{queue && (
					<div className="text-sm text-muted-foreground mb-2">
						Ticket:{" "}
						<span className="font-mono font-semibold">{queue.ticket_no}</span>
						{queue.patient_name && <> — {queue.patient_name}</>}
					</div>
				)}
				<Form {...form}>
					<form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
						<FormField
							control={form.control}
							name="destination_service_id"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Destination Service</FormLabel>
									<Select
										onValueChange={field.onChange}
										defaultValue={field.value}
									>
										<FormControl>
											<SelectTrigger>
												<SelectValue placeholder="Select service" />
											</SelectTrigger>
										</FormControl>
										<SelectContent>
											{services.map((svc) => (
												<SelectItem key={svc.id} value={svc.id}>
													{svc.code} — {svc.name}
												</SelectItem>
											))}
										</SelectContent>
									</Select>
									<FormMessage />
								</FormItem>
							)}
						/>
						<div className="pt-4 flex justify-end">
							<Button type="submit" disabled={isLoading}>
								{isLoading && (
									<Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
								)}
								Forward
							</Button>
						</div>
					</form>
				</Form>
			</DialogContent>
		</Dialog>
	);
}
