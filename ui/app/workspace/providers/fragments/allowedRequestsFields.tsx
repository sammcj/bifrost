"use client";

import { FormControl, FormField, FormItem, FormLabel } from "@/components/ui/form";
import { Switch } from "@/components/ui/switch";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { BaseProvider } from "@/lib/types/config";
import { isRequestTypeDisabled } from "@/lib/utils/validation";
import { Control } from "react-hook-form";

interface AllowedRequestsFieldsProps {
	control: Control<any>;
	namePrefix?: string;
	providerType?: BaseProvider;
}

const REQUEST_TYPES = [
	{ key: "list_models", label: "List Models" },
	{ key: "text_completion", label: "Text Completion" },
	{ key: "chat_completion", label: "Chat Completion" },
	{ key: "chat_completion_stream", label: "Chat Completion Stream" },
	{ key: "responses", label: "Responses" },
	{ key: "responses_stream", label: "Responses Stream" },
	{ key: "embedding", label: "Embedding" },
	{ key: "speech", label: "Speech" },
	{ key: "speech_stream", label: "Speech Stream" },
	{ key: "transcription", label: "Transcription" },
	{ key: "transcription_stream", label: "Transcription Stream" },
];

export function AllowedRequestsFields({ control, namePrefix = "allowed_requests", providerType }: AllowedRequestsFieldsProps) {
	const leftColumn = REQUEST_TYPES.slice(0, 6);
	const rightColumn = REQUEST_TYPES.slice(6);

	const renderRequestField = (requestType: { key: string; label: string }) => {
		const isDisabled = isRequestTypeDisabled(providerType, requestType.key);

		return (
			<FormField
				key={requestType.key}
				control={control}
				name={`${namePrefix}.${requestType.key}`}
				render={({ field }) => (
					<FormItem
						className={`flex flex-row items-center justify-between rounded-lg border p-3 ${isDisabled ? "bg-muted/30 opacity-60" : ""}`}
					>
						<div className="space-y-0.5">
							<FormLabel className={isDisabled ? "cursor-not-allowed" : ""}>{requestType.label}</FormLabel>
						</div>
						<FormControl>
							{isDisabled ? (
								<TooltipProvider>
									<Tooltip>
										<TooltipTrigger asChild>
											<div>
												<Switch checked={field.value} disabled={true} size="md" />
											</div>
										</TooltipTrigger>
										<TooltipContent>
											<p>Not supported by {providerType}</p>
										</TooltipContent>
									</Tooltip>
								</TooltipProvider>
							) : (
								<Switch checked={field.value} onCheckedChange={field.onChange} size="md" />
							)}
						</FormControl>
					</FormItem>
				)}
			/>
		);
	};

	return (
		<div className="space-y-4">
			<div>
				<div className="text-sm font-medium">Allowed Request Types</div>
				<p className="text-muted-foreground text-xs">Select which request types this custom provider can handle</p>
			</div>

			<div className="grid grid-cols-2 gap-4">
				<div className="space-y-3">{leftColumn.map(renderRequestField)}</div>
				<div className="space-y-3">{rightColumn.map(renderRequestField)}</div>
			</div>
		</div>
	);
}
