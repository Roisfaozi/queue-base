import { Suspense } from "react";
import Link from "next/link";
import { AuthLayoutShell } from "~/components/auth/auth-layout-shell";
import { ResetPasswordForm } from "~/components/auth/reset-password-form";
import { Icon } from "~/components/shared/icon";

export default function ResetPassword() {
	return (
		<AuthLayoutShell
			title="Reset password"
			description="Enter your new password below"
			brandingTitle="Almost there!"
			brandingDescription="Once you update your password, you'll be able to sign in and access your workspace."
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
			<Suspense
				fallback={
					<div className="flex justify-center p-8">
						<Icon name="Loader" className="text-primary h-6 w-6 animate-spin" />
					</div>
				}
			>
				<ResetPasswordForm />
			</Suspense>
		</AuthLayoutShell>
	);
}
