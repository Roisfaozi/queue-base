"use client";

import { useSettings } from "./settings-context";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "~/components/ui/card";
import { Label } from "~/components/ui/label";
import { Input } from "~/components/ui/input";
import { LogoUpload } from "~/components/upload";

export function GeneralSettingsCard() {
	const { name, setName, logoAssetId, setLogoAssetId, currentOrganization } =
		useSettings();

	if (!currentOrganization) return null;

	return (
		<Card>
			<CardHeader>
				<CardTitle>General Information</CardTitle>
				<CardDescription>
					The display name, brand logo, and identity of your organization.
				</CardDescription>
			</CardHeader>
			<CardContent className="space-y-6">
				<div className="grid gap-2">
					<Label>Organization Logo</Label>
					<LogoUpload
						scope="organization"
						value={logoAssetId}
						name={name}
						onChange={(url) => setLogoAssetId(url)}
						title="Upload Organization Logo"
						description="PNG, JPG, WEBP, or GIF (<15MB direct upload, ≥15MB via TUS)"
					/>
				</div>

				<div className="grid gap-2">
					<Label htmlFor="name">Organization Name</Label>
					<Input
						id="name"
						value={name}
						onChange={(e) => setName(e.target.value)}
						placeholder="e.g. Acme Corp"
					/>
				</div>
				<div className="grid gap-2">
					<Label htmlFor="slug">Organization Slug (URL)</Label>
					<Input
						id="slug"
						value={currentOrganization.slug}
						disabled
						className="bg-muted font-mono text-xs"
					/>
					<p className="text-muted-foreground text-[10px] italic">
						Slug cannot be changed after creation.
					</p>
				</div>
			</CardContent>
		</Card>
	);
}
