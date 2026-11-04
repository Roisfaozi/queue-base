"use client";

import { useState, useMemo } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
import { Icon } from "~/components/shared/icon";
import { authApi } from "~/lib/api/auth";

const resetPasswordSchema = z
	.object({
		token: z.string().trim().min(1, "Reset token is required."),
		password: z.string().min(8, "Password must be at least 8 characters long."),
		confirmPassword: z.string().min(1, "Please confirm your password."),
	})
	.refine((data) => data.password === data.confirmPassword, {
		message: "Passwords do not match.",
		path: ["confirmPassword"],
	});

type ResetPasswordFormValues = z.infer<typeof resetPasswordSchema>;

export function ResetPasswordForm() {
	const router = useRouter();
	const searchParams = useSearchParams();
	const tokenFromUrl = useMemo(
		() => searchParams.get("token") || "",
		[searchParams],
	);

	const [isLoading, setIsLoading] = useState(false);
	const [isSuccess, setIsSuccess] = useState(false);
	const [showPassword, setShowPassword] = useState(false);
	const [showConfirmPassword, setShowConfirmPassword] = useState(false);

	const {
		register,
		handleSubmit,
		watch,
		formState: { errors },
	} = useForm<ResetPasswordFormValues>({
		resolver: zodResolver(resetPasswordSchema),
		defaultValues: {
			token: tokenFromUrl,
			password: "",
			confirmPassword: "",
		},
		values: tokenFromUrl
			? { token: tokenFromUrl, password: "", confirmPassword: "" }
			: undefined,
	});

	const passwordValue = watch("password") || "";
	const confirmPasswordValue = watch("confirmPassword") || "";

	const hasMinLength = passwordValue.length >= 8;
	const passwordsMatch =
		passwordValue.length > 0 &&
		confirmPasswordValue.length > 0 &&
		passwordValue === confirmPasswordValue;

	async function onSubmit(data: ResetPasswordFormValues) {
		setIsLoading(true);
		try {
			await authApi.resetPassword(data.token, data.password);
			setIsSuccess(true);
			toast.success("Password reset successfully!", {
				description: "You can now log in with your new password.",
			});
		} catch (error: any) {
			toast.error(
				error.message ||
					"Failed to reset password. The link or token may be invalid or expired.",
			);
		} finally {
			setIsLoading(false);
		}
	}

	if (isSuccess) {
		return (
			<div className="space-y-6 text-center">
				<div className="bg-primary/10 text-primary mx-auto flex h-14 w-14 items-center justify-center rounded-full">
					<Icon name="CircleCheck" className="h-7 w-7" />
				</div>
				<div className="space-y-2">
					<h2 className="text-xl font-bold tracking-tight">All set!</h2>
					<p className="text-muted-foreground text-sm">
						Your password has been securely reset. You can now log in to your
						account.
					</p>
				</div>
				<Button className="w-full" onClick={() => router.push("/login")}>
					Continue to Login
				</Button>
			</div>
		);
	}

	return (
		<form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
			{/* Token Field (shown only if not present in query params, or as an editable input) */}
			{!tokenFromUrl ? (
				<div className="grid gap-2">
					<Label htmlFor="token">Reset Token</Label>
					<Input
						id="token"
						placeholder="Paste your reset token"
						disabled={isLoading}
						{...register("token")}
					/>
					{errors.token && (
						<p className="text-destructive text-xs">{errors.token.message}</p>
					)}
				</div>
			) : (
				<input type="hidden" {...register("token")} value={tokenFromUrl} />
			)}

			<div className="grid gap-2">
				<Label htmlFor="password">New Password</Label>
				<div className="relative">
					<Input
						id="password"
						type={showPassword ? "text" : "password"}
						placeholder="••••••••"
						autoComplete="new-password"
						disabled={isLoading}
						className="pr-10"
						{...register("password")}
					/>
					<button
						type="button"
						onClick={() => setShowPassword(!showPassword)}
						className="text-muted-foreground hover:text-foreground absolute top-1/2 right-3 -translate-y-1/2 transition-colors"
						aria-label={showPassword ? "Hide password" : "Show password"}
					>
						<Icon name={showPassword ? "EyeOff" : "Eye"} className="h-4 w-4" />
					</button>
				</div>
				{errors.password && (
					<p className="text-destructive text-xs">{errors.password.message}</p>
				)}
			</div>

			<div className="grid gap-2">
				<Label htmlFor="confirmPassword">Confirm New Password</Label>
				<div className="relative">
					<Input
						id="confirmPassword"
						type={showConfirmPassword ? "text" : "password"}
						placeholder="••••••••"
						autoComplete="new-password"
						disabled={isLoading}
						className="pr-10"
						{...register("confirmPassword")}
					/>
					<button
						type="button"
						onClick={() => setShowConfirmPassword(!showConfirmPassword)}
						className="text-muted-foreground hover:text-foreground absolute top-1/2 right-3 -translate-y-1/2 transition-colors"
						aria-label={showConfirmPassword ? "Hide password" : "Show password"}
					>
						<Icon
							name={showConfirmPassword ? "EyeOff" : "Eye"}
							className="h-4 w-4"
						/>
					</button>
				</div>
				{errors.confirmPassword && (
					<p className="text-destructive text-xs">
						{errors.confirmPassword.message}
					</p>
				)}
			</div>

			{/* Password requirement hints */}
			{passwordValue.length > 0 && (
				<div className="bg-muted/40 border-border space-y-1.5 rounded-lg border p-3 text-xs">
					<div className="flex items-center gap-2">
						<Icon
							name={hasMinLength ? "Check" : "Circle"}
							className={`h-3.5 w-3.5 ${
								hasMinLength ? "text-emerald-500" : "text-muted-foreground"
							}`}
						/>
						<span
							className={
								hasMinLength
									? "text-foreground font-medium"
									: "text-muted-foreground"
							}
						>
							At least 8 characters
						</span>
					</div>
					{confirmPasswordValue.length > 0 && (
						<div className="flex items-center gap-2">
							<Icon
								name={passwordsMatch ? "Check" : "Circle"}
								className={`h-3.5 w-3.5 ${
									passwordsMatch ? "text-emerald-500" : "text-muted-foreground"
								}`}
							/>
							<span
								className={
									passwordsMatch
										? "text-foreground font-medium"
										: "text-muted-foreground"
								}
							>
								Passwords match
							</span>
						</div>
					)}
				</div>
			)}

			<Button type="submit" className="mt-2 w-full" disabled={isLoading}>
				{isLoading ? (
					<>
						<Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
						Updating password...
					</>
				) : (
					"Update Password"
				)}
			</Button>
		</form>
	);
}
