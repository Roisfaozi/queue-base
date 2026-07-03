"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { useDashboardShell } from "~/app/[locale]/dashboard/_components/dashboard-shell-context";
import { Icon } from "~/components/shared/icon";
import { Button } from "~/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "~/components/ui/card";
import {
	Form,
	FormControl,
	FormDescription,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "~/components/ui/form";
import { Input } from "~/components/ui/input";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "~/components/ui/select";
import { branchesApi, qmsClientsApi, type Branch } from "~/lib/api/qms";

const clientTypes = ["caller", "signage", "scanner", "kiosk"] as const;

const clientSchema = z.object({
	branch_id: z.string().min(1, "Branch is required."),
	client_type: z.enum(clientTypes),
	name: z.string().min(1, "Name is required.").max(255),
});

const credentialSchema = z.object({
	client_id: z.string().min(1, "Client ID is required."),
	api_key: z.string().min(8, "API key must be at least 8 characters."),
});

type ClientFormValues = z.infer<typeof clientSchema>;
type CredentialFormValues = z.infer<typeof credentialSchema>;

type CreatedClient = {
	id: string;
	branch_id: string;
	client_type: string;
	name: string;
};

type CreatedCredential = {
	id: string;
	client_id: string;
	expires_at?: number;
};

export function QMSClientsContent() {
	const { currentOrganization } = useDashboardShell();
	const [branches, setBranches] = useState<Branch[]>([]);
	const [createdClient, setCreatedClient] = useState<CreatedClient | null>(
		null,
	);
	const [createdCredential, setCreatedCredential] =
		useState<CreatedCredential | null>(null);
	const [isClientLoading, setIsClientLoading] = useState(false);
	const [isCredentialLoading, setIsCredentialLoading] = useState(false);

	const clientForm = useForm<ClientFormValues>({
		resolver: zodResolver(clientSchema),
		defaultValues: {
			branch_id: "",
			client_type: "caller",
			name: "",
		},
	});

	const credentialForm = useForm<CredentialFormValues>({
		resolver: zodResolver(credentialSchema),
		defaultValues: {
			client_id: "",
			api_key: "",
		},
	});

	const fetchBranches = useCallback(async () => {
		try {
			const response = await branchesApi.getAll();
			setBranches(
				(response.data || []).filter((branch) => branch.status === "active"),
			);
		} catch (error: any) {
			toast.error(error.message || "Failed to load branches");
		}
	}, []);

	useEffect(() => {
		if (currentOrganization) {
			fetchBranches();
		}
	}, [currentOrganization, fetchBranches]);

	useEffect(() => {
		if (createdClient?.id) {
			credentialForm.setValue("client_id", createdClient.id);
		}
	}, [createdClient?.id, credentialForm]);

	const selectedBranchId = clientForm.watch("branch_id");
	const selectedBranch = useMemo(
		() => branches.find((branch) => branch.id === selectedBranchId),
		[branches, selectedBranchId],
	);

	async function onCreateClient(data: ClientFormValues) {
		setIsClientLoading(true);
		setCreatedCredential(null);
		try {
			const response = await qmsClientsApi.create(data);
			setCreatedClient(response.data);
			toast.success("QMS client created");
		} catch (error: any) {
			toast.error(error.message || "Failed to create QMS client");
		} finally {
			setIsClientLoading(false);
		}
	}

	async function onCreateCredential(data: CredentialFormValues) {
		setIsCredentialLoading(true);
		try {
			const response = await qmsClientsApi.createCredential(data);
			setCreatedCredential(response.data);
			toast.success("QMS credential created");
		} catch (error: any) {
			toast.error(error.message || "Failed to create QMS credential");
		} finally {
			setIsCredentialLoading(false);
		}
	}

	if (!currentOrganization) return null;

	return (
		<div className="space-y-6">
			<div>
				<h2 className="text-2xl font-bold tracking-tight">QMS Clients</h2>
				<p className="text-muted-foreground">
					Create device clients and credentials for caller, signage, scanner, or
					kiosk surfaces.
				</p>
			</div>

			<div className="grid gap-6 lg:grid-cols-2">
				<Card>
					<CardHeader>
						<CardTitle>Create QMS Client</CardTitle>
						<CardDescription>
							Bind one QMS device/client to an active branch.
						</CardDescription>
					</CardHeader>
					<CardContent>
						<Form {...clientForm}>
							<form
								onSubmit={clientForm.handleSubmit(onCreateClient)}
								className="space-y-4"
							>
								<FormField
									control={clientForm.control}
									name="branch_id"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Branch</FormLabel>
											<Select
												onValueChange={field.onChange}
												value={field.value}
											>
												<FormControl>
													<SelectTrigger>
														<SelectValue placeholder="Select branch" />
													</SelectTrigger>
												</FormControl>
												<SelectContent>
													{branches.length === 0 ? (
														<SelectItem value="__no_branches__" disabled>
															No active branches
														</SelectItem>
													) : (
														branches.map((branch) => (
															<SelectItem key={branch.id} value={branch.id}>
																{branch.code} — {branch.name}
															</SelectItem>
														))
													)}
												</SelectContent>
											</Select>
											<FormDescription>
												{selectedBranch
													? `Selected: ${selectedBranch.name}`
													: "Only active branches are shown."}
											</FormDescription>
											<FormMessage />
										</FormItem>
									)}
								/>

								<FormField
									control={clientForm.control}
									name="client_type"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Client Type</FormLabel>
											<Select
												onValueChange={field.onChange}
												value={field.value}
											>
												<FormControl>
													<SelectTrigger>
														<SelectValue placeholder="Select client type" />
													</SelectTrigger>
												</FormControl>
												<SelectContent>
													{clientTypes.map((type) => (
														<SelectItem key={type} value={type}>
															{type}
														</SelectItem>
													))}
												</SelectContent>
											</Select>
											<FormMessage />
										</FormItem>
									)}
								/>

								<FormField
									control={clientForm.control}
									name="name"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Name</FormLabel>
											<FormControl>
												<Input placeholder="Main Caller Counter 1" {...field} />
											</FormControl>
											<FormMessage />
										</FormItem>
									)}
								/>

								<Button type="submit" disabled={isClientLoading}>
									{isClientLoading && (
										<Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
									)}
									Create Client
								</Button>
							</form>
						</Form>
					</CardContent>
				</Card>

				<Card>
					<CardHeader>
						<CardTitle>Create Credential</CardTitle>
						<CardDescription>
							Generate or register an API key for a QMS client. Store the secret
							outside this app.
						</CardDescription>
					</CardHeader>
					<CardContent className="space-y-4">
						{createdClient && (
							<div className="rounded-md border bg-muted/30 p-3 text-sm">
								<p className="font-medium">Last created client</p>
								<p className="mt-1 font-mono text-xs">{createdClient.id}</p>
								<p className="mt-1 text-muted-foreground">
									{createdClient.client_type} — {createdClient.name}
								</p>
							</div>
						)}

						<Form {...credentialForm}>
							<form
								onSubmit={credentialForm.handleSubmit(onCreateCredential)}
								className="space-y-4"
							>
								<FormField
									control={credentialForm.control}
									name="client_id"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Client ID</FormLabel>
											<FormControl>
												<Input placeholder="QMS client UUID" {...field} />
											</FormControl>
											<FormMessage />
										</FormItem>
									)}
								/>

								<FormField
									control={credentialForm.control}
									name="api_key"
									render={({ field }) => (
										<FormItem>
											<FormLabel>API Key</FormLabel>
											<FormControl>
												<Input
													placeholder="Paste generated secret"
													{...field}
												/>
											</FormControl>
											<FormDescription>
												Backend stores only a credential hash; keep the
												plaintext key in your device setup notes.
											</FormDescription>
											<FormMessage />
										</FormItem>
									)}
								/>

								<Button type="submit" disabled={isCredentialLoading}>
									{isCredentialLoading && (
										<Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
									)}
									Create Credential
								</Button>
							</form>
						</Form>

						{createdCredential && (
							<div className="rounded-md border bg-muted/30 p-3 text-sm">
								<p className="font-medium">Credential created</p>
								<p className="mt-1 font-mono text-xs">{createdCredential.id}</p>
								<p className="mt-1 text-muted-foreground">
									Client: {createdCredential.client_id}
								</p>
							</div>
						)}
					</CardContent>
				</Card>
			</div>
		</div>
	);
}
