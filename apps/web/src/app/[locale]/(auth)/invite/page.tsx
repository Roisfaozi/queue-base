import { redirect } from "next/navigation";

interface InviteIndexProps {
	params: Promise<{ locale: string }>;
	searchParams: Promise<{ token?: string }>;
}

export default async function InviteIndexPage({
	params,
	searchParams,
}: InviteIndexProps) {
	const { locale } = await params;
	const { token } = await searchParams;

	if (token) {
		redirect(`/${locale}/invite/${encodeURIComponent(token)}`);
	}

	redirect(`/${locale}/login`);
}
