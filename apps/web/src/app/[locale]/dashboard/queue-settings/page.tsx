import { redirect } from "next/navigation";

type Props = {
	params: Promise<{ locale: string }>;
};

export default async function QueueSettingsLegacyPage({ params }: Props) {
	const { locale } = await params;
	redirect(`/${locale}/dashboard/queue-config`);
}
