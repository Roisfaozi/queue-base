"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "~/components/ui/button";
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
import { AvatarUpload } from "~/components/upload";
import { toast } from "sonner";
import { Icon } from "~/components/shared/icon";
import { usersApi } from "~/lib/api/users";
import { useAuthStore } from "~/stores/use-auth-store";

const profileSchema = z.object({
	name: z.string().min(2, "Name must be at least 2 characters."),
	email: z.string().email(),
	username: z.string().min(3),
});

type ProfileFormValues = z.infer<typeof profileSchema>;

interface ProfileFormProps {
	user: any;
}

export function ProfileForm({ user }: ProfileFormProps) {
	const [isLoading, setIsLoading] = useState(false);

	const form = useForm<ProfileFormValues>({
		resolver: zodResolver(profileSchema),
		defaultValues: {
			name: user?.name || "",
			email: user?.email || "",
			username: user?.username || "",
		},
	});

	async function onSubmit(data: ProfileFormValues) {
		setIsLoading(true);
		try {
			await usersApi.updateMe({ name: data.name });
			toast.success("Profile updated successfully");
		} catch (error: any) {
			toast.error(error.message || "Failed to update profile");
		} finally {
			setIsLoading(false);
		}
	}

	return (
		<Form {...form}>
			<form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
				{/* Avatar Section */}
				<div className="flex flex-col sm:flex-row sm:items-center gap-4 sm:gap-6 border-b pb-6">
					<AvatarUpload
						value={user?.avatar_url}
						name={user?.name || user?.username}
						size="lg"
						onChange={(url) => {
							if (user) {
								useAuthStore.getState().setUser({
									...user,
									avatar_url: url,
								});
							}
						}}
					/>
					<div className="space-y-1">
						<h4 className="text-sm font-semibold text-foreground">
							Foto Profil
						</h4>
						<p className="text-xs text-muted-foreground">
							Klik atau seret gambar baru untuk memperbarui avatar akun Anda.
						</p>
						<p className="text-[11px] text-muted-foreground/80">
							Mendukung JPG, PNG, WEBP, atau GIF (&lt;15MB upload biasa,
							&ge;15MB via TUS).
						</p>
					</div>
				</div>

				<FormField
					control={form.control}
					name="username"
					render={({ field }) => (
						<FormItem>
							<FormLabel>Username</FormLabel>
							<FormControl>
								<Input {...field} disabled />
							</FormControl>
							<FormDescription>
								This is your public display name. It can be your real name or a
								pseudonym.
							</FormDescription>
							<FormMessage />
						</FormItem>
					)}
				/>

				<FormField
					control={form.control}
					name="name"
					render={({ field }) => (
						<FormItem>
							<FormLabel>Full Name</FormLabel>
							<FormControl>
								<Input placeholder="John Doe" {...field} />
							</FormControl>
							<FormMessage />
						</FormItem>
					)}
				/>

				<FormField
					control={form.control}
					name="email"
					render={({ field }) => (
						<FormItem>
							<FormLabel>Email</FormLabel>
							<FormControl>
								<Input placeholder="john@example.com" {...field} disabled />
							</FormControl>
							<FormDescription>
								Email cannot be changed directly. Contact support if needed.
							</FormDescription>
							<FormMessage />
						</FormItem>
					)}
				/>

				<Button type="submit" disabled={isLoading}>
					{isLoading && (
						<Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
					)}
					Update Profile
				</Button>
			</form>
		</Form>
	);
}
