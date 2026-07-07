import Link from "next/link";
import { notFound } from "next/navigation";
import type { ReactNode } from "react";
import { Badge } from "~/components/ui/badge";
import { Button, buttonVariants } from "~/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "~/components/ui/card";
import { Checkbox } from "~/components/ui/checkbox";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
import { Progress } from "~/components/ui/progress";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "~/components/ui/select";
import { Separator } from "~/components/ui/separator";
import { Skeleton } from "~/components/ui/skeleton";
import { Switch } from "~/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "~/components/ui/tabs";
import { Textarea } from "~/components/ui/textarea";
import { cn } from "~/lib/utils";

type ShowcaseEntry = {
	slug: string;
	title: string;
	description: string;
	category: string;
	level:
		| "Base UI"
		| "Shared Utility"
		| "Composed Component"
		| "Feature Section"
		| "Page Section";
	path: string;
	variants?: string[];
};

const showcaseEntries: ShowcaseEntry[] = [
	// Base UI primitives
	{
		slug: "accordion",
		title: "Accordion",
		description: "Expandable disclosure primitive.",
		category: "Disclosure",
		level: "Base UI",
		path: "apps/web/src/components/ui/accordion.tsx",
		variants: ["single", "multiple", "nested section"],
	},
	{
		slug: "alert-dialog",
		title: "Alert Dialog",
		description: "Confirmation dialog for destructive or blocking actions.",
		category: "Overlay",
		level: "Base UI",
		path: "apps/web/src/components/ui/alert-dialog.tsx",
		variants: ["default", "destructive", "with footer actions"],
	},
	{
		slug: "alert",
		title: "Alert",
		description: "Inline feedback and callout block.",
		category: "Feedback",
		level: "Base UI",
		path: "apps/web/src/components/ui/alert.tsx",
		variants: ["default", "destructive", "with icon"],
	},
	{
		slug: "aspect-ratio",
		title: "Aspect Ratio",
		description: "Fixed-ratio media container.",
		category: "Layout",
		level: "Base UI",
		path: "apps/web/src/components/ui/aspect-ratio.tsx",
		variants: ["16:9", "1:1", "card media"],
	},
	{
		slug: "avatar",
		title: "Avatar",
		description: "User image and fallback initials.",
		category: "Data Display",
		level: "Base UI",
		path: "apps/web/src/components/ui/avatar.tsx",
		variants: ["image", "fallback", "stack"],
	},
	{
		slug: "badge",
		title: "Badge",
		description: "Small semantic label.",
		category: "Data Display",
		level: "Base UI",
		path: "apps/web/src/components/ui/badge.tsx",
		variants: [
			"default",
			"secondary",
			"destructive",
			"success",
			"warning",
			"info",
			"subtle",
		],
	},
	{
		slug: "breadcrumb",
		title: "Breadcrumb",
		description: "Hierarchical navigation trail.",
		category: "Navigation",
		level: "Base UI",
		path: "apps/web/src/components/ui/breadcrumb.tsx",
		variants: ["simple", "nested", "current page"],
	},
	{
		slug: "button",
		title: "Button",
		description: "Action button primitive.",
		category: "Actions",
		level: "Base UI",
		path: "apps/web/src/components/ui/button.tsx",
		variants: [
			"default",
			"secondary",
			"outline",
			"ghost",
			"link",
			"destructive",
			"magic",
			"outline-solid",
			"sm",
			"lg",
			"icon",
		],
	},
	{
		slug: "calendar",
		title: "Calendar",
		description: "Date picker calendar surface.",
		category: "Input",
		level: "Base UI",
		path: "apps/web/src/components/ui/calendar.tsx",
		variants: ["single", "range", "disabled days"],
	},
	{
		slug: "card",
		title: "Card",
		description: "Generic section container.",
		category: "Layout",
		level: "Base UI",
		path: "apps/web/src/components/ui/card.tsx",
		variants: ["header", "content", "footer", "media card", "stat card"],
	},
	{
		slug: "carousel",
		title: "Carousel",
		description: "Horizontal slide container.",
		category: "Layout",
		level: "Base UI",
		path: "apps/web/src/components/ui/carousel.tsx",
		variants: ["basic", "cards", "controls"],
	},
	{
		slug: "chart",
		title: "Chart",
		description: "Chart container and tooltip helpers.",
		category: "Data Display",
		level: "Base UI",
		path: "apps/web/src/components/ui/chart.tsx",
		variants: ["line", "bar", "tooltip"],
	},
	{
		slug: "checkbox",
		title: "Checkbox",
		description: "Boolean input primitive.",
		category: "Input",
		level: "Base UI",
		path: "apps/web/src/components/ui/checkbox.tsx",
		variants: ["unchecked", "checked", "disabled"],
	},
	{
		slug: "collapsible",
		title: "Collapsible",
		description: "Show/hide content primitive.",
		category: "Disclosure",
		level: "Base UI",
		path: "apps/web/src/components/ui/collapsible.tsx",
		variants: ["closed", "open", "settings row"],
	},
	{
		slug: "command",
		title: "Command",
		description: "Searchable command list surface.",
		category: "Navigation",
		level: "Base UI",
		path: "apps/web/src/components/ui/command.tsx",
		variants: ["palette", "empty", "grouped"],
	},
	{
		slug: "context-menu",
		title: "Context Menu",
		description: "Right-click action menu.",
		category: "Overlay",
		level: "Base UI",
		path: "apps/web/src/components/ui/context-menu.tsx",
		variants: ["default", "submenu", "checkbox item"],
	},
	{
		slug: "dialog",
		title: "Dialog",
		description: "Modal overlay primitive.",
		category: "Overlay",
		level: "Base UI",
		path: "apps/web/src/components/ui/dialog.tsx",
		variants: ["trigger", "form", "wide"],
	},
	{
		slug: "drawer",
		title: "Drawer",
		description: "Mobile drawer overlay.",
		category: "Overlay",
		level: "Base UI",
		path: "apps/web/src/components/ui/drawer.tsx",
		variants: ["bottom", "form", "navigation"],
	},
	{
		slug: "dropdown-menu",
		title: "Dropdown Menu",
		description: "Button-triggered menu.",
		category: "Overlay",
		level: "Base UI",
		path: "apps/web/src/components/ui/dropdown-menu.tsx",
		variants: ["items", "checkbox", "radio", "shortcut"],
	},
	{
		slug: "form",
		title: "Form",
		description: "React Hook Form wrappers.",
		category: "Input",
		level: "Base UI",
		path: "apps/web/src/components/ui/form.tsx",
		variants: ["field", "message", "description"],
	},
	{
		slug: "hover-card",
		title: "Hover Card",
		description: "Hover-triggered detail popover.",
		category: "Overlay",
		level: "Base UI",
		path: "apps/web/src/components/ui/hover-card.tsx",
		variants: ["profile", "metadata"],
	},
	{
		slug: "input-otp",
		title: "Input OTP",
		description: "One-time-code segmented input.",
		category: "Input",
		level: "Base UI",
		path: "apps/web/src/components/ui/input-otp.tsx",
		variants: ["6 digit", "grouped", "disabled"],
	},
	{
		slug: "input",
		title: "Input",
		description: "Text input primitive.",
		category: "Input",
		level: "Base UI",
		path: "apps/web/src/components/ui/input.tsx",
		variants: ["text", "number", "disabled", "with label"],
	},
	{
		slug: "label",
		title: "Label",
		description: "Accessible label primitive.",
		category: "Input",
		level: "Base UI",
		path: "apps/web/src/components/ui/label.tsx",
		variants: ["default", "required", "with control"],
	},
	{
		slug: "menubar",
		title: "Menubar",
		description: "Horizontal application menu.",
		category: "Navigation",
		level: "Base UI",
		path: "apps/web/src/components/ui/menubar.tsx",
		variants: ["root", "submenu", "shortcut"],
	},
	{
		slug: "navigation-menu",
		title: "Navigation Menu",
		description: "Desktop navigation menu system.",
		category: "Navigation",
		level: "Base UI",
		path: "apps/web/src/components/ui/navigation-menu.tsx",
		variants: ["list", "content", "viewport"],
	},
	{
		slug: "pagination",
		title: "Pagination",
		description: "Page navigation controls.",
		category: "Navigation",
		level: "Base UI",
		path: "apps/web/src/components/ui/pagination.tsx",
		variants: ["previous next", "numbers", "ellipsis"],
	},
	{
		slug: "popover",
		title: "Popover",
		description: "Anchored floating content.",
		category: "Overlay",
		level: "Base UI",
		path: "apps/web/src/components/ui/popover.tsx",
		variants: ["default", "form", "filter"],
	},
	{
		slug: "progress",
		title: "Progress",
		description: "Linear progress indicator.",
		category: "Feedback",
		level: "Base UI",
		path: "apps/web/src/components/ui/progress.tsx",
		variants: ["low", "medium", "complete"],
	},
	{
		slug: "radio-group",
		title: "Radio Group",
		description: "Mutually-exclusive choice group.",
		category: "Input",
		level: "Base UI",
		path: "apps/web/src/components/ui/radio-group.tsx",
		variants: ["vertical", "settings", "disabled"],
	},
	{
		slug: "resizable",
		title: "Resizable",
		description: "Resizable panel layout.",
		category: "Layout",
		level: "Base UI",
		path: "apps/web/src/components/ui/resizable.tsx",
		variants: ["two panel", "three panel", "vertical"],
	},
	{
		slug: "scroll-area",
		title: "Scroll Area",
		description: "Styled scroll container.",
		category: "Layout",
		level: "Base UI",
		path: "apps/web/src/components/ui/scroll-area.tsx",
		variants: ["list", "card", "horizontal"],
	},
	{
		slug: "select",
		title: "Select",
		description: "Controlled select primitive.",
		category: "Input",
		level: "Base UI",
		path: "apps/web/src/components/ui/select.tsx",
		variants: ["default", "disabled", "grouped"],
	},
	{
		slug: "separator",
		title: "Separator",
		description: "Visual divider.",
		category: "Layout",
		level: "Base UI",
		path: "apps/web/src/components/ui/separator.tsx",
		variants: ["horizontal", "vertical"],
	},
	{
		slug: "sheet",
		title: "Sheet",
		description: "Side-panel overlay.",
		category: "Overlay",
		level: "Base UI",
		path: "apps/web/src/components/ui/sheet.tsx",
		variants: ["right", "left", "form"],
	},
	{
		slug: "sidebar-ui",
		title: "Sidebar UI",
		description: "Sidebar primitive set.",
		category: "Layout",
		level: "Base UI",
		path: "apps/web/src/components/ui/sidebar.tsx",
		variants: ["expanded", "collapsed", "inset"],
	},
	{
		slug: "skeleton",
		title: "Skeleton",
		description: "Loading placeholder.",
		category: "Feedback",
		level: "Base UI",
		path: "apps/web/src/components/ui/skeleton.tsx",
		variants: ["text", "card", "table"],
	},
	{
		slug: "slider",
		title: "Slider",
		description: "Numeric range input.",
		category: "Input",
		level: "Base UI",
		path: "apps/web/src/components/ui/slider.tsx",
		variants: ["single", "range"],
	},
	{
		slug: "sonner",
		title: "Sonner",
		description: "Toast renderer wrapper.",
		category: "Feedback",
		level: "Base UI",
		path: "apps/web/src/components/ui/sonner.tsx",
		variants: ["success", "error", "loading"],
	},
	{
		slug: "switch",
		title: "Switch",
		description: "Binary toggle input.",
		category: "Input",
		level: "Base UI",
		path: "apps/web/src/components/ui/switch.tsx",
		variants: ["on", "off", "disabled"],
	},
	{
		slug: "table",
		title: "Table",
		description: "Data table primitives.",
		category: "Data Display",
		level: "Base UI",
		path: "apps/web/src/components/ui/table.tsx",
		variants: ["basic", "striped", "actions"],
	},
	{
		slug: "tabs",
		title: "Tabs",
		description: "Tabbed content primitive.",
		category: "Navigation",
		level: "Base UI",
		path: "apps/web/src/components/ui/tabs.tsx",
		variants: ["two tabs", "three tabs", "card tabs"],
	},
	{
		slug: "textarea",
		title: "Textarea",
		description: "Multiline text input.",
		category: "Input",
		level: "Base UI",
		path: "apps/web/src/components/ui/textarea.tsx",
		variants: ["default", "disabled", "composer"],
	},
	{
		slug: "toast",
		title: "Toast",
		description: "Toast content primitives.",
		category: "Feedback",
		level: "Base UI",
		path: "apps/web/src/components/ui/toast.tsx",
		variants: ["default", "destructive", "action"],
	},
	{
		slug: "toaster",
		title: "Toaster",
		description: "Toast viewport host.",
		category: "Feedback",
		level: "Base UI",
		path: "apps/web/src/components/ui/toaster.tsx",
		variants: ["viewport", "stack"],
	},
	{
		slug: "toggle-group",
		title: "Toggle Group",
		description: "Grouped toggle controls.",
		category: "Input",
		level: "Base UI",
		path: "apps/web/src/components/ui/toggle-group.tsx",
		variants: ["single", "multiple", "icons"],
	},
	{
		slug: "toggle",
		title: "Toggle",
		description: "Pressed/unpressed control.",
		category: "Input",
		level: "Base UI",
		path: "apps/web/src/components/ui/toggle.tsx",
		variants: ["default", "outline", "disabled"],
	},
	{
		slug: "tooltip",
		title: "Tooltip",
		description: "Hover/focus helper text.",
		category: "Overlay",
		level: "Base UI",
		path: "apps/web/src/components/ui/tooltip.tsx",
		variants: ["top", "right", "icon"],
	},
	// Shared utilities
	{
		slug: "brand-icons",
		title: "Brand Icons",
		description: "Reusable brand SVG/icon helpers.",
		category: "Brand",
		level: "Shared Utility",
		path: "apps/web/src/components/shared/brand-icons.tsx",
		variants: ["github", "google", "brand row"],
	},
	{
		slug: "copy-button",
		title: "Copy Button",
		description: "Clipboard copy action.",
		category: "Actions",
		level: "Shared Utility",
		path: "apps/web/src/components/shared/copy-button.tsx",
		variants: ["idle", "copied", "icon only"],
	},
	{
		slug: "density-switcher",
		title: "Density Switcher",
		description: "Comfort/compact density switch control.",
		category: "Settings",
		level: "Shared Utility",
		path: "apps/web/src/components/shared/density-switcher.tsx",
		variants: ["comfort", "compact"],
	},
	{
		slug: "density-toggle",
		title: "Density Toggle",
		description: "Quick density toggle button.",
		category: "Settings",
		level: "Shared Utility",
		path: "apps/web/src/components/shared/density-toggle.tsx",
		variants: ["button", "menu item"],
	},
	{
		slug: "empty-state",
		title: "Empty State",
		description: "Generic empty data state.",
		category: "Feedback",
		level: "Shared Utility",
		path: "apps/web/src/components/shared/empty-state.tsx",
		variants: ["default", "with action", "compact"],
	},
	{
		slug: "empty-states",
		title: "Empty States",
		description: "Preset empty states for common domains.",
		category: "Feedback",
		level: "Shared Utility",
		path: "apps/web/src/components/shared/empty-states.tsx",
		variants: ["users", "projects", "queues"],
	},
	{
		slug: "global-search",
		title: "Global Search",
		description: "Dashboard command/search trigger.",
		category: "Navigation",
		level: "Shared Utility",
		path: "apps/web/src/components/shared/global-search.tsx",
		variants: ["comfort", "compact"],
	},
	{
		slug: "go-back",
		title: "Go Back",
		description: "Back navigation helper.",
		category: "Navigation",
		level: "Shared Utility",
		path: "apps/web/src/components/shared/go-back.tsx",
		variants: ["button", "link"],
	},
	{
		slug: "icon",
		title: "Icon",
		description: "Density-aware Lucide icon wrapper.",
		category: "Iconography",
		level: "Shared Utility",
		path: "apps/web/src/components/shared/icon.tsx",
		variants: ["sm", "md", "lg", "xl"],
	},
	{
		slug: "icons",
		title: "Icons",
		description: "Shared icon collection.",
		category: "Iconography",
		level: "Shared Utility",
		path: "apps/web/src/components/shared/icons.tsx",
		variants: ["set", "semantic"],
	},
	{
		slug: "locale-toggler",
		title: "Locale Toggler",
		description: "Language switch control.",
		category: "Settings",
		level: "Shared Utility",
		path: "apps/web/src/components/shared/locale-toggler.tsx",
		variants: ["compact", "dropdown"],
	},
	{
		slug: "logout-button",
		title: "Logout Button",
		description: "Session sign-out control.",
		category: "Actions",
		level: "Shared Utility",
		path: "apps/web/src/components/shared/logout-button.tsx",
		variants: ["button", "menu item"],
	},
	{
		slug: "search-input",
		title: "Search Input",
		description: "Reusable search input surface.",
		category: "Input",
		level: "Shared Utility",
		path: "apps/web/src/components/shared/search-input.tsx",
		variants: ["default", "with icon", "compact"],
	},
	{
		slug: "skeletons",
		title: "Skeleton Presets",
		description: "Reusable loading skeleton patterns.",
		category: "Feedback",
		level: "Shared Utility",
		path: "apps/web/src/components/shared/skeletons.tsx",
		variants: ["card", "table", "form"],
	},
	{
		slug: "smart-form-field",
		title: "Smart Form Field",
		description: "Higher-level form field composition.",
		category: "Input",
		level: "Shared Utility",
		path: "apps/web/src/components/shared/smart-form-field.tsx",
		variants: ["text", "select", "textarea"],
	},
	{
		slug: "theme-toggle",
		title: "Theme Toggle",
		description: "Light/dark theme toggle.",
		category: "Settings",
		level: "Shared Utility",
		path: "apps/web/src/components/shared/theme-toggle.tsx",
		variants: ["light", "dark", "system"],
	},
	// Composed components
	{
		slug: "auth-layout-shell",
		title: "Auth Layout Shell",
		description: "Auth page layout wrapper.",
		category: "Auth",
		level: "Composed Component",
		path: "apps/web/src/components/auth/auth-layout-shell.tsx",
		variants: ["login", "register"],
	},
	{
		slug: "login-form",
		title: "Login Form",
		description: "Sign-in form composition.",
		category: "Auth",
		level: "Composed Component",
		path: "apps/web/src/components/auth/login-form.tsx",
		variants: ["default", "loading", "error"],
	},
	{
		slug: "register-form",
		title: "Register Form",
		description: "Registration form composition.",
		category: "Auth",
		level: "Composed Component",
		path: "apps/web/src/components/auth/register-form.tsx",
		variants: ["default", "validation", "success"],
	},
	{
		slug: "activity-chart",
		title: "Activity Chart",
		description: "Dashboard analytics chart.",
		category: "Dashboard",
		level: "Composed Component",
		path: "apps/web/src/components/dashboard/activity-chart.tsx",
		variants: ["line", "bar", "empty"],
	},
	{
		slug: "create-organization-modal",
		title: "Create Organization Modal",
		description: "Organization creation modal flow.",
		category: "Dashboard",
		level: "Composed Component",
		path: "apps/web/src/components/dashboard/create-organization-modal.tsx",
		variants: ["form", "loading", "error"],
	},
	{
		slug: "email-verification-banner",
		title: "Email Verification Banner",
		description: "Account verification callout.",
		category: "Dashboard",
		level: "Composed Component",
		path: "apps/web/src/components/dashboard/email-verification-banner.tsx",
		variants: ["warning", "sent", "verified"],
	},
	{
		slug: "kpi-card",
		title: "KPI Card",
		description: "Dashboard metric card.",
		category: "Dashboard",
		level: "Composed Component",
		path: "apps/web/src/components/dashboard/kpi-card.tsx",
		variants: ["comfort", "compact", "trend"],
	},
	{
		slug: "notification-center",
		title: "Notification Center",
		description: "Header notification dropdown.",
		category: "Dashboard",
		level: "Composed Component",
		path: "apps/web/src/components/dashboard/notification-center.tsx",
		variants: ["empty", "unread", "list"],
	},
	{
		slug: "organization-switcher",
		title: "Organization Switcher",
		description: "Current organization selector.",
		category: "Dashboard",
		level: "Composed Component",
		path: "apps/web/src/components/dashboard/organization-switcher.tsx",
		variants: ["selected", "empty", "create"],
	},
	{
		slug: "presence-avatar-stack",
		title: "Presence Avatar Stack",
		description: "Online users avatar cluster.",
		category: "Dashboard",
		level: "Composed Component",
		path: "apps/web/src/components/dashboard/presence-avatar-stack.tsx",
		variants: ["few", "many", "offline"],
	},
	{
		slug: "profile-form",
		title: "Profile Form",
		description: "User profile edit form.",
		category: "Dashboard",
		level: "Composed Component",
		path: "apps/web/src/components/dashboard/profile-form.tsx",
		variants: ["default", "saving", "avatar"],
	},
	{
		slug: "security-form",
		title: "Security Form",
		description: "Security settings form.",
		category: "Dashboard",
		level: "Composed Component",
		path: "apps/web/src/components/dashboard/security-form.tsx",
		variants: ["password", "2fa", "sessions"],
	},
	{
		slug: "user-nav",
		title: "User Nav",
		description: "User menu in dashboard header.",
		category: "Dashboard",
		level: "Composed Component",
		path: "apps/web/src/components/dashboard/user-nav.tsx",
		variants: ["profile", "settings", "logout"],
	},
	// Layout and page sections
	{
		slug: "announcement-banner",
		title: "Announcement Banner",
		description: "Top announcement strip.",
		category: "Layout",
		level: "Page Section",
		path: "apps/web/src/components/layout/announcement-banner.tsx",
		variants: ["info", "cta", "dismissible"],
	},
	{
		slug: "cancel-confirm-modal",
		title: "Cancel Confirm Modal",
		description: "Cancel confirmation modal.",
		category: "Layout",
		level: "Page Section",
		path: "apps/web/src/components/layout/cancel-confirm-modal.tsx",
		variants: ["default", "destructive"],
	},
	{
		slug: "footer",
		title: "Footer",
		description: "Site footer section.",
		category: "Layout",
		level: "Page Section",
		path: "apps/web/src/components/layout/footer.tsx",
		variants: ["marketing", "minimal"],
	},
	{
		slug: "header",
		title: "Header",
		description: "Marketing site header.",
		category: "Layout",
		level: "Page Section",
		path: "apps/web/src/components/layout/header.tsx",
		variants: ["desktop", "mobile"],
	},
	{
		slug: "image-upload-modal",
		title: "Image Upload Modal",
		description: "Image upload modal workflow.",
		category: "Layout",
		level: "Page Section",
		path: "apps/web/src/components/layout/image-upload-modal.tsx",
		variants: ["empty", "preview", "uploading"],
	},
	{
		slug: "login-modal",
		title: "Login Modal",
		description: "Auth modal overlay.",
		category: "Layout",
		level: "Page Section",
		path: "apps/web/src/components/layout/login-modal.tsx",
		variants: ["login", "register"],
	},
	{
		slug: "navbar",
		title: "Navbar",
		description: "Dashboard top navigation.",
		category: "Layout",
		level: "Page Section",
		path: "apps/web/src/components/layout/navbar.tsx",
		variants: ["comfort", "compact", "actions"],
	},
	{
		slug: "sidebar",
		title: "Sidebar",
		description: "Dashboard side navigation.",
		category: "Layout",
		level: "Page Section",
		path: "apps/web/src/components/layout/sidebar.tsx",
		variants: ["expanded", "compact", "active item"],
	},
	{
		slug: "dashboard-header",
		title: "Dashboard Header",
		description: "Dashboard header composition.",
		category: "Layout",
		level: "Page Section",
		path: "apps/web/src/components/layout/dashboard/header.tsx",
		variants: ["default", "compact"],
	},
	// Landing sections
	{
		slug: "faq",
		title: "FAQ",
		description: "Landing FAQ accordion section.",
		category: "Landing",
		level: "Feature Section",
		path: "apps/web/src/components/landing/faq.tsx",
		variants: ["accordion", "two column"],
	},
	{
		slug: "features",
		title: "Features",
		description: "Landing feature grid.",
		category: "Landing",
		level: "Feature Section",
		path: "apps/web/src/components/landing/features.tsx",
		variants: ["grid", "cards"],
	},
	{
		slug: "hero",
		title: "Hero",
		description: "Landing hero section.",
		category: "Landing",
		level: "Feature Section",
		path: "apps/web/src/components/landing/hero.tsx",
		variants: ["center", "split"],
	},
	{
		slug: "open-source",
		title: "Open Source",
		description: "Open-source positioning section.",
		category: "Landing",
		level: "Feature Section",
		path: "apps/web/src/components/landing/open-source.tsx",
		variants: ["default", "stats"],
	},
	{
		slug: "pricing",
		title: "Pricing",
		description: "Pricing tiers section.",
		category: "Landing",
		level: "Feature Section",
		path: "apps/web/src/components/landing/pricing.tsx",
		variants: ["two tier", "three tier"],
	},
	{
		slug: "tech-stack",
		title: "Tech Stack",
		description: "Technology logo/stack section.",
		category: "Landing",
		level: "Feature Section",
		path: "apps/web/src/components/landing/tech-stack.tsx",
		variants: ["logo grid", "marquee"],
	},
	{
		slug: "testimonials",
		title: "Testimonials",
		description: "Customer testimonial cards.",
		category: "Landing",
		level: "Feature Section",
		path: "apps/web/src/components/landing/testimonials.tsx",
		variants: ["cards", "marquee"],
	},
	// Magic UI
	{
		slug: "bento-grid",
		title: "Bento Grid",
		description: "Asymmetric bento layout helper.",
		category: "Magic UI",
		level: "Shared Utility",
		path: "apps/web/src/components/magicui/bento-grid.tsx",
		variants: ["two column", "dashboard"],
	},
	{
		slug: "marquee",
		title: "Marquee",
		description: "Scrolling marquee helper.",
		category: "Magic UI",
		level: "Shared Utility",
		path: "apps/web/src/components/magicui/marquee.tsx",
		variants: ["horizontal", "vertical"],
	},
	{
		slug: "retro-grid",
		title: "Retro Grid",
		description: "Decorative retro grid background.",
		category: "Magic UI",
		level: "Shared Utility",
		path: "apps/web/src/components/magicui/retro-grid.tsx",
		variants: ["hero", "section"],
	},
	{
		slug: "word-pull-up",
		title: "Word Pull Up",
		description: "Animated word reveal text.",
		category: "Magic UI",
		level: "Shared Utility",
		path: "apps/web/src/components/magicui/word-pull-up.tsx",
		variants: ["heading", "caption"],
	},
];

const hierarchyOrder: ShowcaseEntry["level"][] = [
	"Base UI",
	"Shared Utility",
	"Composed Component",
	"Feature Section",
	"Page Section",
];
const sharedCard = "rounded-xl border bg-card p-6 shadow-sm";

function titleToSlug(title: string) {
	return title
		.toLowerCase()
		.replace(/[^a-z0-9]+/g, "-")
		.replace(/(^-|-$)/g, "");
}

function getCategoryGroups() {
	return hierarchyOrder
		.map((level) => ({
			level,
			entries: showcaseEntries.filter((entry) => entry.level === level),
		}))
		.filter((group) => group.entries.length > 0);
}

function ComponentPreview({
	entry,
	variant,
}: {
	entry: ShowcaseEntry;
	variant?: string;
}) {
	const isVariantPreview = !!variant && variant !== "default";

	if (entry.slug === "button") {
		return (
			<div className="flex flex-wrap items-center gap-2">
				<Button>Call Next</Button>
				<Button variant="secondary">Transfer</Button>
				<Button variant="outline">Hold Queue</Button>
				<Button variant="ghost">Pause</Button>
				<Button variant="destructive">Cancel Ticket</Button>
				<Button size="sm">Quick</Button>
				<Button size="lg">Start Serving</Button>
			</div>
		);
	}

	if (["input", "search-input", "smart-form-field"].includes(entry.slug)) {
		return (
			<div className="grid max-w-sm gap-2">
				<Label htmlFor={`${entry.slug}-preview`}>Search queue / ticket</Label>
				<Input
					id={`${entry.slug}-preview`}
					placeholder="A-012, Rina, or Customer Service"
				/>
				<p className="text-muted-foreground text-xs">
					Search by ticket number, customer name, or service.
				</p>
			</div>
		);
	}

	if (entry.slug === "select") {
		return (
			<Select defaultValue="one">
				<SelectTrigger className="w-[220px]">
					<SelectValue placeholder="Choose service" />
				</SelectTrigger>
				<SelectContent>
					<SelectItem value="one">Customer Service</SelectItem>
					<SelectItem value="two">Payment Counter</SelectItem>
					<SelectItem value="three">Priority Lane</SelectItem>
				</SelectContent>
			</Select>
		);
	}

	if (
		[
			"checkbox",
			"switch",
			"density-switcher",
			"density-toggle",
			"theme-toggle",
		].includes(entry.slug)
	) {
		return (
			<div className="grid gap-4 sm:grid-cols-2">
				<div className="flex items-center gap-3 rounded-lg border bg-background p-4">
					<Checkbox id={`${entry.slug}-check`} defaultChecked />
					<div className="grid gap-1">
						<Label htmlFor={`${entry.slug}-check`}>Auto call next ticket</Label>
						<p className="text-muted-foreground text-xs">
							Notify operator when queue advances.
						</p>
					</div>
				</div>
				<div className="flex items-center justify-between rounded-lg border bg-background p-4">
					<div className="grid gap-1">
						<Label htmlFor={`${entry.slug}-switch`}>
							Publish to waiting room
						</Label>
						<p className="text-muted-foreground text-xs">
							Show current ticket status on display screen.
						</p>
					</div>
					<Switch id={`${entry.slug}-switch`} defaultChecked />
				</div>
			</div>
		);
	}

	if (["progress", "skeleton", "skeletons"].includes(entry.slug)) {
		return (
			<div className="grid max-w-sm gap-3">
				<div className="flex items-center justify-between text-sm">
					<span className="font-medium">Serving queue</span>
					<span className="text-muted-foreground">64% complete</span>
				</div>
				<Progress value={64} />
				<Skeleton className="h-4 w-full" />
				<Skeleton className="h-4 w-3/4" />
			</div>
		);
	}

	if (["tabs", "navigation-menu", "menubar"].includes(entry.slug)) {
		return (
			<Tabs defaultValue="one" className="max-w-md">
				<TabsList>
					<TabsTrigger value="one">Queue</TabsTrigger>
					<TabsTrigger value="two">Serve</TabsTrigger>
					<TabsTrigger value="three">History</TabsTrigger>
				</TabsList>
				<TabsContent value="one" className="rounded-md border p-3">
					Queue 18 waiting, 4 being served.
				</TabsContent>
				<TabsContent value="two" className="rounded-md border p-3">
					Current operator: Rina, Counter 03.
				</TabsContent>
				<TabsContent value="three" className="rounded-md border p-3">
					Last called: A-011, A-010, A-009.
				</TabsContent>
			</Tabs>
		);
	}

	if (entry.slug === "badge") {
		return (
			<div className="flex flex-wrap gap-2">
				<Badge>Waiting</Badge>
				<Badge variant="secondary">Serving</Badge>
				<Badge variant="destructive">Missed</Badge>
				<Badge variant="success">Completed</Badge>
				<Badge variant="warning">Priority</Badge>
				<Badge variant="outline">A-012</Badge>
			</div>
		);
	}

	if (entry.slug === "textarea") {
		return (
			<div className="grid max-w-sm gap-2">
				<Label htmlFor="preview-textarea">Serve note</Label>
				<Textarea
					id="preview-textarea"
					placeholder="Customer asks for wheelchair access and priority queue"
					className="min-h-20 resize-none"
				/>
			</div>
		);
	}

	if (["dialog", "alert-dialog", "sheet"].includes(entry.slug)) {
		return (
			<Card className="shadow-none">
				<CardHeader className="pb-3">
					<CardTitle className="text-base">Call ticket A-012</CardTitle>
					<CardDescription>
						Customer waits at Branch Jakarta Selatan, Counter 03.
					</CardDescription>
				</CardHeader>
				<CardContent className="flex flex-wrap gap-2">
					<Button size="sm">Confirm call</Button>
					<Button size="sm" variant="outline">
						Send to counter
					</Button>
				</CardContent>
			</Card>
		);
	}

	if (["popover", "dropdown-menu", "context-menu"].includes(entry.slug)) {
		return (
			<Card className="shadow-none">
				<CardHeader className="pb-3">
					<CardTitle className="text-sm font-medium">Menu preview</CardTitle>
				</CardHeader>
				<CardContent className="grid gap-3 text-sm">
					<p className="cursor-pointer hover:underline">View ticket detail</p>
					<p className="cursor-pointer hover:underline">
						Move to priority lane
					</p>
					<p className="cursor-pointer text-destructive hover:underline">
						Cancel visit
					</p>
				</CardContent>
			</Card>
		);
	}

	if (["table", "data-table", "command", "command-menu"].includes(entry.slug)) {
		return (
			<div className="overflow-hidden rounded-lg border">
				<div className="grid grid-cols-[1.2fr_1fr_1fr_0.8fr] gap-0 border-b bg-muted/40 px-4 py-3 text-xs font-medium uppercase tracking-wider">
					<span>Ticket</span>
					<span>Service</span>
					<span>Customer</span>
					<span>Status</span>
				</div>
				<div className="grid grid-cols-[1.2fr_1fr_1fr_0.8fr] gap-0 px-4 py-3 text-sm">
					<span>A-012</span>
					<span>Customer Service</span>
					<span>Rina</span>
					<span>
						<Badge>Waiting</Badge>
					</span>
				</div>
				<div className="grid grid-cols-[1.2fr_1fr_1fr_0.8fr] gap-0 border-t px-4 py-3 text-sm">
					<span>A-013</span>
					<span>Payment Counter</span>
					<span>Andi</span>
					<span>
						<Badge variant="secondary">Serving</Badge>
					</span>
				</div>
			</div>
		);
	}

	if (["calendar", "date-picker"].includes(entry.slug)) {
		return (
			<Card className="max-w-sm">
				<CardHeader>
					<CardTitle className="text-base">07 Jul 2026</CardTitle>
					<CardDescription>
						Queue reset day and branch schedule.
					</CardDescription>
				</CardHeader>
				<CardContent className="grid grid-cols-7 gap-2 text-center text-sm">
					{["Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"].map((day) => (
						<div
							key={day}
							className="text-muted-foreground text-xs font-medium"
						>
							{day}
						</div>
					))}
					{[
						"1",
						"2",
						"3",
						"4",
						"5",
						"6",
						"7",
						"8",
						"9",
						"10",
						"11",
						"12",
						"13",
						"14",
					].map((day) => (
						<div
							key={day}
							className={cn(
								"rounded-md border px-0 py-2",
								day === "7" ? "bg-primary text-primary-foreground" : "",
							)}
						>
							{day}
						</div>
					))}
				</CardContent>
			</Card>
		);
	}

	if (["avatar", "presence-avatar-stack", "user-nav"].includes(entry.slug)) {
		return (
			<div className="flex items-center gap-4">
				<div className="flex -space-x-2">
					<div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary text-primary-foreground text-sm font-semibold">
						RN
					</div>
					<div className="flex h-10 w-10 items-center justify-center rounded-full bg-muted text-sm font-semibold">
						AS
					</div>
					<div className="flex h-10 w-10 items-center justify-center rounded-full bg-muted/80 text-sm font-semibold">
						+2
					</div>
				</div>
				<div>
					<p className="font-medium">Rina Putri</p>
					<p className="text-muted-foreground text-sm">
						Operator Counter 03 • online
					</p>
				</div>
			</div>
		);
	}

	if (entry.slug === "card" || entry.slug === "kpi-card") {
		return (
			<div className="grid gap-4 md:grid-cols-3">
				<Card className="md:col-span-2">
					<CardHeader>
						<CardTitle className="text-base">Today queue</CardTitle>
						<CardDescription>
							Live tickets for Branch Jakarta Selatan.
						</CardDescription>
					</CardHeader>
					<CardContent className="flex items-end justify-between gap-4">
						<div>
							<p className="text-3xl font-semibold">128 waiting</p>
							<p className="text-muted-foreground text-sm">
								36 served, 4 missed
							</p>
						</div>
						<Badge variant="secondary">+12% vs yesterday</Badge>
					</CardContent>
				</Card>
				<Card>
					<CardHeader>
						<CardTitle className="text-base">Avg serve time</CardTitle>
					</CardHeader>
					<CardContent>
						<p className="text-2xl font-semibold">4m 12s</p>
						<p className="text-muted-foreground text-sm">
							Fastest at counter 02
						</p>
					</CardContent>
				</Card>
			</div>
		);
	}

	if (entry.slug === "accordion") {
		return (
			<div className="grid gap-1">
				<div className="rounded-md border bg-muted/30 px-3 py-2 text-sm font-medium">
					Service Info
				</div>
				<div className="rounded-md border bg-card px-3 py-2 text-sm">
					Customer waiting at Counter 03 since 10:32.
				</div>
				<div className="rounded-md border bg-card px-3 py-2 text-sm">
					Ticket A-012, Priority Normal.
				</div>
			</div>
		);
	}

	if (["alert", "toast", "sonner", "toaster"].includes(entry.slug)) {
		return (
			<div className="grid max-w-sm gap-3">
				<div className="rounded-md border-l-4 border-l-emerald-500 bg-muted/40 px-4 py-3 text-sm">
					<span className="font-medium">Ticket called</span> — A-012 to Counter
					03.
				</div>
				<div className="rounded-md border-l-4 border-l-destructive bg-muted/40 px-4 py-3 text-sm">
					<span className="font-medium">Skip detected</span> — A-011 missed 3
					calls.
				</div>
			</div>
		);
	}

	if (entry.slug === "aspect-ratio") {
		return (
			<div className="flex gap-4">
				<div className="flex aspect-video w-32 items-center justify-center rounded-lg border bg-muted/40 text-muted-foreground text-xs">
					16:9
				</div>
				<div className="flex aspect-square w-12 items-center justify-center rounded-lg border bg-muted/40 text-muted-foreground text-xs">
					1:1
				</div>
			</div>
		);
	}

	if (["separator", "scroll-area", "resizable"].includes(entry.slug)) {
		return (
			<div className="flex max-w-sm items-center gap-4 rounded-lg border bg-card p-4">
				<div className="flex-1 text-sm">
					<span className="font-medium">Ticket A-012</span>
					<br />
					<span className="text-muted-foreground text-xs">Waiting</span>
				</div>
				<Separator orientation="vertical" className="h-10" />
				<div className="flex-1 text-sm">
					<span className="font-medium">Counter 03</span>
					<br />
					<span className="text-muted-foreground text-xs">Serving A-011</span>
				</div>
			</div>
		);
	}

	if (["carousel", "slider"].includes(entry.slug)) {
		return (
			<div className="flex max-w-sm gap-2 overflow-hidden rounded-lg border bg-muted/20 p-3">
				<div className="flex-1 rounded-md border bg-card p-3 text-center text-xs font-medium">
					CS
				</div>
				<div className="flex-1 rounded-md border bg-card p-3 text-center text-xs font-medium">
					Payment
				</div>
				<div className="flex-1 rounded-md border bg-card p-3 text-center text-xs font-medium">
					Priority
				</div>
			</div>
		);
	}

	if (
		[
			"chart",
			"label",
			"radio-group",
			"toggle",
			"toggle-group",
			"tooltip",
			"form",
			"input-otp",
			"pagination",
			"sidebar",
			"collapsible",
		].includes(entry.slug)
	) {
		return (
			<div className="grid max-w-sm gap-3 rounded-lg border bg-card p-4">
				<div className="flex items-center gap-3">
					<div className="flex h-10 w-10 items-center justify-center rounded-md bg-primary/10 text-sm">
						CS
					</div>
					<div className="grid gap-0.5">
						<p className="text-sm font-medium">{entry.title}</p>
						<p className="text-muted-foreground text-xs">
							Ready for {entry.level} interaction
						</p>
					</div>
					<Badge variant="secondary" className="ml-auto">
						demo
					</Badge>
				</div>
				<Separator />
				<div className="flex justify-between text-sm">
					<span className="text-muted-foreground">Status</span>
					<Badge>active</Badge>
				</div>
			</div>
		);
	}

	if (["Layout", "Dashboard", "Landing"].includes(entry.category)) {
		return (
			<div className="rounded-2xl border bg-gradient-to-br from-primary/5 via-background to-muted/20 p-8">
				<div className="max-w-xl space-y-4">
					<Badge variant="secondary">Branch dashboard</Badge>
					<h3 className="text-2xl font-semibold tracking-tight">
						Queue ops for 07 Jul 2026
					</h3>
					<p className="text-muted-foreground text-sm">
						Display current queue, operator status, and service movement in one
						glance.
					</p>
					<div className="flex flex-wrap gap-2">
						<Button>Open queue</Button>
						<Button variant="outline">View counters</Button>
					</div>
				</div>
			</div>
		);
	}

	return (
		<Card className="max-w-md">
			<CardHeader>
				<div className="flex items-center justify-between gap-3">
					<CardTitle className="text-base">{entry.title}</CardTitle>
					<Badge variant="secondary">{entry.level}</Badge>
				</div>
				<CardDescription>{entry.description}</CardDescription>
			</CardHeader>
			<CardContent>
				<div className="flex flex-wrap gap-2">
					{(entry.variants ?? []).slice(0, 4).map((variant) => (
						<Badge key={variant} variant="outline">
							{variant}
						</Badge>
					))}
				</div>
			</CardContent>
		</Card>
	);
}

function VariantGrid({ entry }: { entry: ShowcaseEntry }) {
	const variants = entry.variants?.length
		? entry.variants
		: ["default", "empty", "loading", "error"];
	return (
		<div className="grid gap-4 md:grid-cols-2">
			{variants.map((variant) => (
				<Card key={variant}>
					<CardHeader>
						<CardTitle className="text-lg">{variant}</CardTitle>
						<CardDescription>
							{entry.title} variant from {entry.path}.
						</CardDescription>
					</CardHeader>
					<CardContent>
						<ComponentPreview entry={entry} variant={variant} />
					</CardContent>
				</Card>
			))}
		</div>
	);
}

export function getShowcaseEntry(slug: string) {
	return showcaseEntries.find((entry) => entry.slug === slug);
}

export function getShowcaseEntries() {
	return showcaseEntries;
}

export function ShowcaseRegistryNav() {
	return (
		<div className="space-y-10">
			{getCategoryGroups().map((group) => (
				<section key={group.level} className="space-y-4">
					<div>
						<h2 className="text-2xl font-semibold tracking-tight">
							{group.level}
						</h2>
						<p className="text-muted-foreground text-sm">
							{group.entries.length} components, ordered from lower-level
							primitives to higher-level sections.
						</p>
					</div>
					<div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
						{group.entries.map((entry) => (
							<div
								key={entry.slug}
								className={cn(
									sharedCard,
									"transition-colors hover:border-primary/40 hover:bg-muted/30",
								)}
							>
								<div className="flex items-start justify-between gap-4">
									<div>
										<p className="text-muted-foreground text-xs uppercase tracking-wider">
											{entry.category}
										</p>
										<h3 className="mt-1 text-lg font-semibold">
											{entry.title}
										</h3>
										<p className="text-muted-foreground mt-2 text-sm">
											{entry.description}
										</p>
									</div>
									<Badge variant="secondary">
										{entry.variants?.length ?? 1}
									</Badge>
								</div>
								<div className="mt-4 rounded-lg border bg-background p-4">
									<ComponentPreview entry={entry} />
								</div>
								<div className="mt-4 flex justify-between gap-3">
									<code className="text-muted-foreground truncate text-xs">
										{entry.path}
									</code>
									<Link
										href={`/dashboard/showcase/${entry.slug}`}
										className={buttonVariants({
											variant: "outline",
											size: "sm",
										})}
									>
										Open
									</Link>
								</div>
							</div>
						))}
					</div>
				</section>
			))}
		</div>
	);
}

export function ShowcasePage({ slug }: { slug?: string }) {
	if (!slug) {
		return (
			<div className="space-y-6">
				<div>
					<h1 className="text-3xl font-bold tracking-tight">
						Component Showcase
					</h1>
					<p className="text-muted-foreground mt-2 max-w-3xl">
						All UI components under <code>apps/web/src/components</code>,
						grouped from base primitives to page-level sections. Each component
						has its own route.
					</p>
				</div>
				<ShowcaseRegistryNav />
			</div>
		);
	}
	const entry = getShowcaseEntry(slug);
	if (!entry) notFound();
	return (
		<div className="space-y-8">
			<div className="flex flex-wrap items-start justify-between gap-4">
				<div>
					<p className="text-muted-foreground text-xs uppercase tracking-wider">
						{entry.level} / {entry.category}
					</p>
					<h1 className="text-3xl font-bold tracking-tight">{entry.title}</h1>
					<p className="text-muted-foreground mt-2 max-w-3xl">
						{entry.description}
					</p>
					<code className="text-muted-foreground mt-2 block text-xs">
						{entry.path}
					</code>
				</div>
				<Link
					href="/dashboard/showcase"
					className={buttonVariants({ variant: "outline" })}
				>
					Back to index
				</Link>
			</div>
			<Card>
				<CardHeader>
					<CardTitle>Preview</CardTitle>
					<CardDescription>
						Representative preview for {entry.title}.
					</CardDescription>
				</CardHeader>
				<CardContent>
					<ComponentPreview entry={entry} />
				</CardContent>
			</Card>
			<VariantGrid entry={entry} />
		</div>
	);
}
