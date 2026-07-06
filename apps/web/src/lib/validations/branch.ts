import { branchActivationSchema } from "@casbin/api-types";

export { branchActivationSchema };

export function validateBranchActivation(payload: {
	address?: string;
	city?: string;
	province?: string;
	phone?: string;
	timezone?: string;
}) {
	return branchActivationSchema.safeParse(payload);
}
