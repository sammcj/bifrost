import { Button } from "@/components/ui/button";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { getErrorMessage, useCreateProviderMutation } from "@/lib/store";
import { BaseProvider, KnownProvider, ModelProviderName } from "@/lib/types/config";
import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { AllowedRequestsFields } from "../fragments/allowedRequestsFields";
import { allowedRequestsSchema } from "@/lib/types/schemas";

const formSchema = z.object({
	name: z.string().min(1),
	baseFormat: z.string().min(1),
	base_url: z.string().min(1, "Base URL is required").url("Must be a valid URL"),
	allowed_requests: allowedRequestsSchema,
});

type FormData = z.infer<typeof formSchema>;

interface Props {
	show: boolean;
	onSave: (id: string) => void;
	onClose: () => void;
}

export default function AddCustomProviderSheet({ show, onClose, onSave }: Props) {
	const [addProvider, { isLoading: isAddingProvider }] = useCreateProviderMutation();
	const form = useForm<FormData>({
		resolver: zodResolver(formSchema),
		defaultValues: {
			name: "",
			baseFormat: "",
			base_url: "",
			allowed_requests: {
				text_completion: true,
				chat_completion: true,
				chat_completion_stream: true,
				responses: true,
				responses_stream: true,
				embedding: true,
				speech: true,
				speech_stream: true,
				transcription: true,
				transcription_stream: true,
				list_models: true,
			},
		},
	});

	useEffect(() => {
		if (show) {
			form.clearErrors();
		}
	}, [show]);

	const onSubmit = (data: FormData) => {
		addProvider({
			provider: data.name as ModelProviderName,
			custom_provider_config: {
				base_provider_type: data.baseFormat as KnownProvider,
				allowed_requests: data.allowed_requests,
			},
			network_config: {
				base_url: data.base_url,
				default_request_timeout_in_seconds: 30,
				max_retries: 0,
				retry_backoff_initial: 500,
				retry_backoff_max: 5000,
			},
			keys: [],
		})
			.unwrap()
			.then((provider) => {
				onSave(provider.name);
				form.reset();
			})
			.catch((err) => {
				toast.error("Failed to add provider", {
					description: getErrorMessage(err),
				});
			});
	};

	return (
		<Sheet open={show} onOpenChange={(open) => !open && onClose()}>
			<SheetContent className="custom-scrollbar dark:bg-card bg-white py-4 min-w-[600px] flex flex-col">
				<SheetHeader>
					<SheetTitle>Add Custom Provider</SheetTitle>
					<SheetDescription>Enter the details of your custom provider.</SheetDescription>
				</SheetHeader>
				<Form {...form}>
					<form onSubmit={form.handleSubmit(onSubmit)} className="flex flex-col flex-1 overflow-hidden">
						<div className="space-y-4 px-4 flex-1 overflow-y-auto custom-scrollbar">
							<FormField
								control={form.control}
								name="name"
								render={({ field }) => (
									<FormItem className="flex flex-col gap-3">
										<FormLabel className="text-right">Name</FormLabel>
										<div className="col-span-3">
											<FormControl>
												<Input placeholder="Name" {...field} />
											</FormControl>
											<FormMessage />
										</div>
									</FormItem>
								)}
							/>
							<FormField
								control={form.control}
								name="baseFormat"
								render={({ field }) => (
									<FormItem className="flex flex-col gap-3">
										<FormLabel>Base Format</FormLabel>
										<div>
											<FormControl>
												<Select onValueChange={field.onChange} defaultValue={field.value}>
													<SelectTrigger className="w-full">
														<SelectValue placeholder="Select base format" />
													</SelectTrigger>
													<SelectContent>
														<SelectItem value="openai">OpenAI</SelectItem>
														<SelectItem value="anthropic">Anthropic</SelectItem>
														<SelectItem value="gemini">Gemini</SelectItem>
														<SelectItem value="cohere">Cohere</SelectItem>
														<SelectItem value="bedrock">AWS Bedrock</SelectItem>
													</SelectContent>
												</Select>
											</FormControl>
											<FormMessage />
										</div>
									</FormItem>
								)}
							/>
							<FormField
								control={form.control}
								name="base_url"
								render={({ field }) => (
									<FormItem className="flex flex-col gap-3">
										<FormLabel>Base URL</FormLabel>
										<div>
											<FormControl>
												<Input placeholder={"https://api.your-provider.com"} {...field} value={field.value || ""} />
											</FormControl>
											<FormMessage />
										</div>
									</FormItem>
								)}
							/>
							{/* Allowed Requests Configuration */}
							<AllowedRequestsFields 
								control={form.control} 
								providerType={form.watch("baseFormat") as BaseProvider}
							/>
						</div>
						<SheetFooter className="flex flex-row gap-2 mt-4 px-4 pt-4">
							<div className="ml-auto flex flex-row gap-2">
								<Button type="button" variant="outline" onClick={onClose}>
									Cancel
								</Button>
								<Button type="submit" isLoading={isAddingProvider}>Add</Button>
							</div>
						</SheetFooter>
					</form>
				</Form>
			</SheetContent>
		</Sheet>
	);
}
