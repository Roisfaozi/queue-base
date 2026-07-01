"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { memo } from "react";
import { useDashboardShell } from "~/app/[locale]/dashboard/_components/dashboard-shell-context";
import { buttonVariants } from "~/components/ui/button";
import {
	Tooltip,
	TooltipContent,
	TooltipProvider,
	TooltipTrigger,
} from "~/components/ui/tooltip";
import { cn } from "~/lib/utils";
import { OrganizationSwitcher } from "../dashboard/organization-switcher";
import { Icon } from "../shared/icon";

// Define Navigation Items
type NavItem = {
	type?: "link";
	title: string;
	href: string;
	iconName: string;
};
type NavSeparator = { type: "separator" };
type NavEntry = NavItem | NavSeparator;

const navItems: NavEntry[] = [
	{
		title: "Dashboard",
		href: "/dashboard",
		iconName: "LayoutDashboard",
	},
	{
		title: "Projects",
		href: "/dashboard/projects",
		iconName: "Folder",
	},
	{
		title: "Users",
		href: "/dashboard/users",
		iconName: "UserSearch",
	},
	{
		title: "Team Members",
		href: "/dashboard/organization/members",
		iconName: "Users",
	},
	{
		title: "Org Settings",
		href: "/dashboard/organization/settings",
		iconName: "Building",
	},
	{
		title: "Roles",
		href: "/dashboard/roles",
		iconName: "Shield",
	},
	{
		title: "Access Matrix",
		href: "/dashboard/access",
		iconName: "Grid3X3",
	},
	{
		title: "Access Rights",
		href: "/dashboard/access-rights",
		iconName: "Key",
	},
	{
		title: "Audit Logs",
		href: "/dashboard/audit",
		iconName: "FileText",
	},
	{ type: "separator" },
	{
		title: "Services",
		href: "/dashboard/services",
		iconName: "Globe",
	},
	{
		title: "Counters",
		href: "/dashboard/counters",
		iconName: "Monitor",
	},
	{
		title: "Queues",
		href: "/dashboard/queues",
		iconName: "ListOrdered",
	},
	{
		title: "Scanner",
		href: "/dashboard/scanner",
		iconName: "Scan",
	},
	{
		title: "Queue Settings",
		href: "/dashboard/queue-settings",
		iconName: "Settings2",
	},
	{ type: "separator" },
	{
		title: "Settings",
		href: "/dashboard/settings",
		iconName: "Settings",
	},
];

export const Sidebar = memo(function Sidebar({
	className,
}: {
	className?: string;
}) {
	const pathname = usePathname();
	const { currentOrganization } = useDashboardShell();

	return (
		<aside
			className={cn(
				"bg-background flex flex-col border-r transition-all duration-300",
				"sticky top-0 h-screen w-[var(--sidebar-width)]",
				className,
			)}
		>
			{/* Header / Logo + Switcher */}
			<div className="flex h-[var(--navbar-height)] items-center gap-2 border-b px-3">
				<Link
					href="/"
					className="flex shrink-0 items-center gap-2 overflow-hidden"
				>
					<Icon name="Command" size="md" className="text-primary" />
				</Link>
				<OrganizationSwitcher />
			</div>

			{/* Navigation */}
			<nav className="flex flex-1 flex-col gap-1 overflow-y-auto p-2">
				{navItems.map((item, index) => {
					if (item.type === "separator") {
						return (
							<div
								key={`sep-${index}`}
								className="bg-border my-2 h-px w-full"
							/>
						);
					}

					const isActive =
						pathname === item.href || pathname.startsWith(`${item.href}/`);

					return (
						<TooltipProvider key={item.href}>
							<Tooltip delayDuration={0}>
								<TooltipTrigger asChild>
									<Link
										href={item.href}
										className={cn(
											buttonVariants({
												variant: isActive ? "secondary" : "ghost",
												size: "default",
											}),
											"w-full justify-start overflow-hidden",
											isActive &&
												"bg-primary/10 text-primary hover:bg-primary/20",
											"[data-density=compact]:justify-center [data-density=compact]:px-0",
										)}
									>
										<Icon
											name={item.iconName as any}
											className={cn(isActive && "text-primary")}
										/>
										<span className="ml-3 truncate [data-density=compact]:hidden">
											{item.title}
										</span>
									</Link>
								</TooltipTrigger>
								<TooltipContent
									side="right"
									className="hidden [data-density=compact]:block"
								>
									{item.title}
								</TooltipContent>
							</Tooltip>
						</TooltipProvider>
					);
				})}
			</nav>

			{/* Footer / User Info Context (Optional) */}
			<div className="border-t p-4 [data-density=compact]:p-2">
				<div className="text-muted-foreground flex flex-col gap-1 text-[10px] font-bold tracking-wider uppercase [data-density=compact]:hidden">
					<span>Active Org</span>
					<span className="text-primary truncate">
						{currentOrganization?.name || "None"}
					</span>
				</div>
			</div>
		</aside>
	);
});
