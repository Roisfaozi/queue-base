"use client";

import { format } from "date-fns";
import { useCallback, useEffect, useState } from "react";
import { EmptyState } from "~/components/shared/empty-state";
import { Icon } from "~/components/shared/icon";
import { Badge } from "~/components/ui/badge";
import { ScrollArea } from "~/components/ui/scroll-area";
import {
	Sheet,
	SheetContent,
	SheetDescription,
	SheetHeader,
	SheetTitle,
} from "~/components/ui/sheet";
import { type Queue, queuesApi, type VisitJourney } from "~/lib/api/qms";

interface QueueDetailSheetProps {
	queue?: Queue | null;
	open: boolean;
	onOpenChange: (open: boolean) => void;
}

export function QueueDetailSheet({
	queue,
	open,
	onOpenChange,
}: QueueDetailSheetProps) {
	const [journeys, setJourneys] = useState<VisitJourney[]>([]);
	const [isLoading, setIsLoading] = useState(false);

	const fetchJourneys = useCallback(async () => {
		if (!queue?.id || !open) return;
		setIsLoading(true);
		try {
			const resp = await queuesApi.getVisitJourneys(queue.id);
			setJourneys(resp.data || []);
		} catch {
			setJourneys([]);
		} finally {
			setIsLoading(false);
		}
	}, [queue?.id, open]);

	useEffect(() => {
		fetchJourneys();
	}, [fetchJourneys]);

	return (
		<Sheet open={open} onOpenChange={onOpenChange}>
			<SheetContent className="sm:max-w-xl w-full flex flex-col gap-0 p-0">
				<div className="p-6 pb-4 border-b">
					<SheetHeader>
						<SheetTitle className="flex items-center gap-2">
							Ticket {queue?.ticket_no}
							<Badge variant="secondary" className="font-mono text-xs">
								#{queue?.queue_no}
							</Badge>
						</SheetTitle>
						<SheetDescription>
							Patient: {queue?.patient_name || "-"}
							{queue?.patient_id ? ` (ID: ${queue.patient_id})` : ""}
						</SheetDescription>
					</SheetHeader>
				</div>

				<ScrollArea className="flex-1">
					<div className="p-6 space-y-6">
						<div>
							<h3 className="text-sm font-medium mb-4">Journey History</h3>
							{isLoading ? (
								<div className="flex justify-center p-8">
									<Icon
										name="Loader"
										className="h-6 w-6 animate-spin text-muted-foreground"
									/>
								</div>
							) : journeys.length === 0 ? (
								<EmptyState
									case="generic"
									title="No journey events"
									description="No activity recorded yet."
								/>
							) : (
								<div className="relative border-l-2 ml-3 space-y-8">
									{journeys.map((j) => (
										<div key={j.id} className="relative pl-6">
											<div className="absolute -left-[9px] top-1 h-4 w-4 rounded-full border-2 border-background bg-primary" />
											<div className="flex flex-col gap-1">
												<div className="flex items-center gap-2">
													<span className="font-semibold capitalize text-sm">
														{j.event_type}
													</span>
													<span className="text-xs text-muted-foreground">
														{format(new Date(j.created_at * 1000), "HH:mm:ss")}
													</span>
												</div>
												{j.payload && (
													<pre className="mt-2 rounded-md bg-muted p-2 text-xs text-muted-foreground overflow-auto">
														{JSON.stringify(JSON.parse(j.payload), null, 2)}
													</pre>
												)}
											</div>
										</div>
									))}
								</div>
							)}
						</div>
					</div>
				</ScrollArea>
			</SheetContent>
		</Sheet>
	);
}
