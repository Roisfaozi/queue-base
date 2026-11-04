import { redirect } from "next/navigation";

interface AcceptInviteProps {
	params: Promise<{ locale: string }>;
	searchParams: Promise<{ token?: string }>;
}

export default async function AcceptInvitePage({
	params,
	searchParams,
}: AcceptInviteProps) {
	const { locale } = await params;
	const { token } = await searchParams;

	if (token) {
		redirect(`/${locale}/invite/${encodeURIComponent(token)}`);
	}

	redirect(`/${locale}/login`);
}
