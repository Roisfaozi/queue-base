"use client";

import { useState } from "react";
import Link from "next/link";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
import { Icon } from "~/components/shared/icon";
import { AuthLayoutShell } from "~/components/auth/auth-layout-shell";
import { authApi } from "~/lib/api/auth";

const forgotPasswordSchema = z.object({
	email: z
		.string()
		.trim()
		.min(1, "Email is required.")
		.email("Please enter a valid email address."),
});

type ForgotPasswordFormValues = z.infer<typeof forgotPasswordSchema>;

export default function ForgotPassword() {
	const [isLoading, setIsLoading] = useState(false);
	const [isSubmitted, setIsSubmitted] = useState(false);
	const [submittedEmail, setSubmittedEmail] = useState("");

	const {
		register,
		handleSubmit,
		formState: { errors },
	} = useForm<ForgotPasswordFormValues>({
		resolver: zodResolver(forgotPasswordSchema),
		defaultValues: {
			email: "",
		},
	});

	async function onSubmit(data: ForgotPasswordFormValues) {
		setIsLoading(true);
		try {
			await authApi.forgotPassword(data.email);
			setSubmittedEmail(data.email);
			setIsSubmitted(true);
			toast.success("Reset link sent", {
				description:
					"If that email exists in our system, you will receive a reset link shortly.",
			});
		} catch (error: any) {
			toast.error(error.message || "Failed to process request");
		} finally {
			setIsLoading(false);
		}
	}

	return (
		<AuthLayoutShell
			title={isSubmitted ? "Check your email" : "Forgot password?"}
			description={
				isSubmitted
					? "We have dispatched a password reset link to your email address."
					: "Enter your email and we'll send you a link to reset your password"
			}
			brandingTitle="Don't Worry!"
			brandingDescription="It happens to the best of us. We'll have you back in your account in no time."
			footer={
				<Link
					href="/login"
					className="text-muted-foreground hover:text-primary flex items-center gap-2 text-sm font-medium transition-colors"
				>
					<Icon name="ChevronLeft" className="h-4 w-4" />
					Back to login
				</Link>
			}
		>
			{isSubmitted ? (
				<div className="space-y-6">
					<div className="border-border bg-muted/40 rounded-lg border p-4 text-center">
						<div className="bg-primary/10 text-primary mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-full">
							<Icon name="Mail" className="h-6 w-6" />
						</div>
						<p className="text-muted-foreground text-sm">
							We sent instructions to:
						</p>
						<p className="text-foreground mt-1 font-semibold break-all">
							{submittedEmail}
						</p>
						<p className="text-muted-foreground mt-3 text-xs">
							Did not receive the email? Check your spam folder or try again
							with another email.
						</p>
					</div>

					<Button
						variant="outline"
						className="w-full"
						onClick={() => setIsSubmitted(false)}
					>
						Resend or use another email
					</Button>
				</div>
			) : (
				<form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
					<div className="grid gap-2">
						<Label htmlFor="email">Email</Label>
						<Input
							id="email"
							placeholder="name@example.com"
							type="email"
							autoCapitalize="none"
							autoComplete="email"
							autoCorrect="off"
							disabled={isLoading}
							{...register("email")}
						/>
						{errors.email && (
							<p className="text-destructive text-xs">{errors.email.message}</p>
						)}
					</div>

					<Button type="submit" className="mt-2 w-full" disabled={isLoading}>
						{isLoading ? (
							<>
								<Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
								Sending link...
							</>
						) : (
							"Send Reset Link"
						)}
					</Button>
				</form>
			)}
		</AuthLayoutShell>
	);
}
