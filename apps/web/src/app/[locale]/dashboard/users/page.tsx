import { UsersProvider } from "./_components/users-context";
import { UsersContent } from "./_components/users-content";

export default function UsersPage() {
	return (
		<UsersProvider>
			<UsersContent />
		</UsersProvider>
	);
}
