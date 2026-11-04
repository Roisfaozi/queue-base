"use client";

import { DashboardHeader } from "~/components/layout/dashboard/header";
import { Sidebar } from "~/components/layout/sidebar";
import { usePresence } from "~/hooks/use-presence";
import { UploadManager, UploadDuplicateDialog } from "~/components/upload";
import { useUploadSideEffects } from "~/lib/upload";
import { toast } from "sonner";

export function DashboardLayoutClient({
	children,
}: {
	children: React.ReactNode;
}) {
	usePresence();

	useUploadSideEffects({
		onToastSuccess: (msg) => toast.success(msg),
		onToastError: (msg) => toast.error(msg),
		onToastWarning: (msg) => toast.warning(msg),
	});

	return (
		<div className="bg-background flex min-h-screen">
			{/* Sidebar */}
			<Sidebar className="z-40 hidden md:flex" />

			{/* Main Area */}
			<div className="flex min-h-screen flex-1 flex-col transition-all">
				<DashboardHeader />

				<main className="flex-1 overflow-y-auto p-[var(--layout-padding)]">
					{children}
				</main>
			</div>

			{/* Global Floating Upload Suite */}
			<UploadManager />
			<UploadDuplicateDialog />
		</div>
	);
}
