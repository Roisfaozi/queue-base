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

// Generated categories for semantic grouping inside submenu
type ShowcaseNavGroup = {
	title: string;
	items: { title: string; href: string; category: string }[];
};

const showcaseNavGroups: ShowcaseNavGroup[] = [
	{
		title: "Base UI",
		items: [
			{ title: "Button", category: "Actions" },
			{ title: "Badge", category: "Data Display" },
			{ title: "Avatar", category: "Data Display" },
			{ title: "Table", category: "Data Display" },
			{ title: "Chart", category: "Data Display" },
			{ title: "Accordion", category: "Disclosure" },
			{ title: "Collapsible", category: "Disclosure" },
			{ title: "Alert", category: "Feedback" },
			{ title: "Skeleton", category: "Feedback" },
			{ title: "Progress", category: "Feedback" },
			{ title: "Toast", category: "Feedback" },
			{ title: "Toaster", category: "Feedback" },
			{ title: "Sonner", category: "Feedback" },
			{ title: "Input", category: "Input" },
			{ title: "Checkbox", category: "Input" },
			{ title: "Radio Group", category: "Input" },
			{ title: "Switch", category: "Input" },
			{ title: "Select", category: "Input" },
			{ title: "Slider", category: "Input" },
			{ title: "Textarea", category: "Input" },
			{ title: "Toggle", category: "Input" },
			{ title: "Toggle Group", category: "Input" },
			{ title: "Form", category: "Input" },
			{ title: "Input OTP", category: "Input" },
			{ title: "Calendar", category: "Input" },
			{ title: "Label", category: "Input" },
			{ title: "Card", category: "Layout" },
			{ title: "Separator", category: "Layout" },
			{ title: "Aspect Ratio", category: "Layout" },
			{ title: "Scroll Area", category: "Layout" },
			{ title: "Resizable", category: "Layout" },
			{ title: "Carousel", category: "Layout" },
			{ title: "Sidebar UI", category: "Layout" },
			{ title: "Breadcrumb", category: "Navigation" },
			{ title: "Command", category: "Navigation" },
			{ title: "Menubar", category: "Navigation" },
			{ title: "Navigation Menu", category: "Navigation" },
			{ title: "Tabs", category: "Navigation" },
			{ title: "Pagination", category: "Navigation" },
			{ title: "Alert Dialog", category: "Overlay" },
			{ title: "Context Menu", category: "Overlay" },
			{ title: "Dialog", category: "Overlay" },
			{ title: "Dropdown Menu", category: "Overlay" },
			{ title: "Hover Card", category: "Overlay" },
			{ title: "Popover", category: "Overlay" },
			{ title: "Sheet", category: "Overlay" },
			{ title: "Drawer", category: "Overlay" },
			{ title: "Tooltip", category: "Overlay" },
		].map((item) => ({
			...item,
			href: `/dashboard/showcase/${item.title
				.toLowerCase()
				.replace(/[^a-z0-9]+/g, "-")
				.replace(/(^-|-$)/g, "")}`,
		})),
	},
	{
		title: "Shared Utility",
		items: [
			{ title: "Copy Button", category: "Actions" },
			{ title: "Logout Button", category: "Actions" },
			{ title: "Brand Icons", category: "Brand" },
			{ title: "Empty State", category: "Feedback" },
			{ title: "Empty States", category: "Feedback" },
			{ title: "Skeleton Presets", category: "Feedback" },
			{ title: "Icon", category: "Iconography" },
			{ title: "Icons", category: "Iconography" },
			{ title: "Search Input", category: "Input" },
			{ title: "Smart Form Field", category: "Input" },
			{ title: "Bento Grid", category: "Magic UI" },
			{ title: "Marquee", category: "Magic UI" },
			{ title: "Retro Grid", category: "Magic UI" },
			{ title: "Word Pull Up", category: "Magic UI" },
			{ title: "Global Search", category: "Navigation" },
			{ title: "Go Back", category: "Navigation" },
			{ title: "Density Switcher", category: "Settings" },
			{ title: "Density Toggle", category: "Settings" },
			{ title: "Locale Toggler", category: "Settings" },
			{ title: "Theme Toggle", category: "Settings" },
		].map((item) => ({
			...item,
			href: `/dashboard/showcase/${item.title
				.toLowerCase()
				.replace(/[^a-z0-9]+/g, "-")
				.replace(/(^-|-$)/g, "")}`,
		})),
	},
	{
		title: "Composed Component",
		items: [
			{ title: "Auth Layout Shell", category: "Auth" },
			{ title: "Login Form", category: "Auth" },
			{ title: "Register Form", category: "Auth" },
			{ title: "Activity Chart", category: "Dashboard" },
			{ title: "Create Organization Modal", category: "Dashboard" },
			{ title: "Email Verification Banner", category: "Dashboard" },
			{ title: "KPI Card", category: "Dashboard" },
			{ title: "Notification Center", category: "Dashboard" },
			{ title: "Organization Switcher", category: "Dashboard" },
			{ title: "Presence Avatar Stack", category: "Dashboard" },
			{ title: "Profile Form", category: "Dashboard" },
			{ title: "Security Form", category: "Dashboard" },
			{ title: "User Nav", category: "Dashboard" },
		].map((item) => ({
			...item,
			href: `/dashboard/showcase/${item.title
				.toLowerCase()
				.replace(/[^a-z0-9]+/g, "-")
				.replace(/(^-|-$)/g, "")}`,
		})),
	},
	{
		title: "Feature Section",
		items: [
			{ title: "FAQ", category: "Landing" },
			{ title: "Features", category: "Landing" },
			{ title: "Hero", category: "Landing" },
			{ title: "Open Source", category: "Landing" },
			{ title: "Pricing", category: "Landing" },
			{ title: "Tech Stack", category: "Landing" },
			{ title: "Testimonials", category: "Landing" },
		].map((item) => ({
			...item,
			href: `/dashboard/showcase/${item.title
				.toLowerCase()
				.replace(/[^a-z0-9]+/g, "-")
				.replace(/(^-|-$)/g, "")}`,
		})),
	},
	{
		title: "Page Section",
		items: [
			{ title: "Announcement Banner", category: "Layout" },
			{ title: "Cancel Confirm Modal", category: "Layout" },
			{ title: "Dashboard Header", category: "Layout" },
			{ title: "Footer", category: "Layout" },
			{ title: "Header", category: "Layout" },
			{ title: "Image Upload Modal", category: "Layout" },
			{ title: "Login Modal", category: "Layout" },
			{ title: "Navbar", category: "Layout" },
			{ title: "Sidebar", category: "Layout" },
		].map((item) => ({
			...item,
			href: `/dashboard/showcase/${item.title
				.toLowerCase()
				.replace(/[^a-z0-9]+/g, "-")
				.replace(/(^-|-$)/g, "")}`,
		})),
	},
];

const navItems: NavEntry[] = [
	{
		title: "Dashboard",
		href: "/dashboard",
		iconName: "LayoutDashboard",
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
		title: "Caller",
		href: "/dashboard/caller",
		iconName: "PhoneCall",
	},
	{
		title: "Signage",
		href: "/dashboard/signage",
		iconName: "MonitorPlay",
	},
	{
		title: "Queue Config",
		href: "/dashboard/queue-config",
		iconName: "Settings2",
	},
	{
		title: "QMS Clients",
		href: "/dashboard/qms-clients",
		iconName: "KeyRound",
	},
	{
		title: "Operator Assignments",
		href: "/dashboard/operator-assignments",
		iconName: "UserCog",
	},
	{
		title: "QMS Setup",
		href: "/dashboard/qms-setup",
		iconName: "ListTodo",
	},
	{
		title: "Branches",
		href: "/dashboard/branches",
		iconName: "GitBranch",
	},
	{ type: "separator" },
	{
		title: "Showcase",
		href: "/dashboard/showcase",
		iconName: "Sparkles",
	},
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
	const normalizedPathname = pathname.replace(/^\/[a-z]{2}(?=\/|$)/, "") || "/";
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
			<nav className="flex flex-1 flex-col gap-1 overflow-y-auto p-2 pb-24">
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
						normalizedPathname === item.href ||
						normalizedPathname.startsWith(`${item.href}/`);
					const isShowcase = item.href === "/dashboard/showcase";

					return (
						<div key={item.href} className="flex flex-col gap-1">
							<TooltipProvider>
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

							{isShowcase && isActive ? (
								<div className="ml-4 border-l border-border pl-3 [data-density=compact]:hidden">
									{showcaseNavGroups.map((group) => {
										const groupActive = group.items.some(
											(child) => normalizedPathname === child.href,
										);

										// Group items by category
										const itemsByCategory = group.items.reduce(
											(acc, item) => {
												if (!acc[item.category]) acc[item.category] = [];
												acc[item.category].push(item);
												return acc;
											},
											{} as Record<string, typeof group.items>,
										);

										return (
											<details
												key={group.title}
												open={groupActive || pathname === item.href}
												className="py-1"
											>
												<summary
													className={cn(
														"flex cursor-pointer list-none items-center gap-2 py-1 text-xs font-semibold tracking-wide uppercase",
														groupActive
															? "text-primary"
															: "text-muted-foreground hover:text-foreground",
													)}
												>
													{group.title}
													<span className="ml-auto rounded-full bg-muted px-1.5 py-0 text-[10px] tabular-nums">
														{group.items.length}
													</span>
												</summary>
												<div className="mt-2 flex flex-col gap-3">
													{Object.entries(itemsByCategory).map(
														([category, items]) => {
															const catActive = items.some(
																(child) => normalizedPathname === child.href,
															);
															return (
																<div
																	key={category}
																	className="flex flex-col gap-0.5"
																>
																	<p
																		className={cn(
																			"flex items-center gap-1 px-2 text-[10px] font-bold tracking-wider uppercase",
																			catActive
																				? "text-primary/70"
																				: "text-muted-foreground/70",
																		)}
																	>
																		{category}
																		<span className="text-muted-foreground/40 text-[9px]">
																			{items.length}
																		</span>
																	</p>
																	{items.map((child) => {
																		const childActive =
																			normalizedPathname === child.href;
																		return (
																			<Link
																				key={child.href}
																				href={child.href}
																				className={cn(
																					"truncate rounded-md px-2 py-1.5 text-sm text-muted-foreground hover:bg-muted hover:text-foreground",
																					childActive &&
																						"bg-primary/10 text-primary",
																				)}
																			>
																				{child.title}
																			</Link>
																		);
																	})}
																</div>
															);
														},
													)}
												</div>
											</details>
										);
									})}
								</div>
							) : null}
						</div>
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
