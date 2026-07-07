import { ShowcasePage } from "../_components/showcase-data";

export default async function ShowcaseComponentPage({
	params,
}: {
	params: Promise<{ slug: string }>;
}) {
	const { slug } = await params;
	return <ShowcasePage slug={slug} />;
}
