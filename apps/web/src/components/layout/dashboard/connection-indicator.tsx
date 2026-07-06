"use client";

import { useWebSocket } from "~/components/shared/providers/websocket-provider";
import { cn } from "~/lib/utils";

export function ConnectionIndicator() {
	const { isConnected } = useWebSocket();

	return (
		<div
			className="flex items-center gap-1.5 text-xs text-muted-foreground"
			title={isConnected ? "Connected" : "Disconnected"}
		>
			<span
				className={cn(
					"inline-block h-2 w-2 rounded-full",
					isConnected ? "bg-emerald-500" : "bg-red-500",
				)}
			/>
			<span>{isConnected ? "Connected" : "Disconnected"}</span>
		</div>
	);
}
