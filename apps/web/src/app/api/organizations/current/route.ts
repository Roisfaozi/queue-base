import { cookies } from "next/headers";
import { NextResponse } from "next/server";

export async function PATCH(request: Request) {
	try {
		const body = (await request.json()) as {
			organizationId?: string;
			organizationSlug?: string;
		};

		if (!body.organizationId || !body.organizationSlug) {
			return NextResponse.json(
				{
					success: false,
					message: "organizationId and organizationSlug are required",
				},
				{ status: 400 },
			);
		}

		const cookieStore = await cookies();
		cookieStore.set("organization_id", body.organizationId, {
			httpOnly: false,
			secure: process.env.NODE_ENV === "production",
			sameSite: "lax",
			path: "/",
		});
		cookieStore.set("organization_slug", body.organizationSlug, {
			httpOnly: false,
			secure: process.env.NODE_ENV === "production",
			sameSite: "lax",
			path: "/",
		});

		return NextResponse.json({ success: true });
	} catch (_error) {
		return NextResponse.json(
			{ success: false, message: "Invalid organization selection payload" },
			{ status: 400 },
		);
	}
}

export async function DELETE() {
	const cookieStore = await cookies();
	cookieStore.delete("organization_id");
	cookieStore.delete("organization_slug");

	return NextResponse.json({ success: true });
}
