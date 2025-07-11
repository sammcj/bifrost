"use client";

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Copy, RefreshCw, ArrowRight, AlertTriangle } from "lucide-react";
import { CodeEditor } from "./ui/code-editor";
import { toast } from "sonner";
import { useState, useMemo } from "react";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Alert, AlertDescription } from "../ui/alert";
import { getExampleBaseUrl } from "@/lib/utils/port";

type Provider = "openai" | "anthropic" | "genai" | "litellm";
type Language = "python" | "typescript";

type Examples = {
	curl: string;
	sdk: {
		[P in Provider]: {
			[L in Language]: string;
		};
	};
};

// Common editor options to reduce duplication
const EDITOR_OPTIONS = {
	scrollBeyondLastLine: false,
	minimap: { enabled: false },
	lineNumbers: "off",
	folding: false,
	lineDecorationsWidth: 0,
	lineNumbersMinChars: 0,
	glyphMargin: false,
} as const;

interface CodeBlockProps {
	code: string;
	language: string;
	onLanguageChange?: (language: string) => void;
	showLanguageSelect?: boolean;
	readonly?: boolean;
}

function CodeBlock({ code, language, onLanguageChange, showLanguageSelect = false, readonly = true }: CodeBlockProps) {
	const copyToClipboard = () => {
		navigator.clipboard.writeText(code);
		toast.success("Copied to clipboard");
	};

	return (
		<div className="relative">
			<div className="absolute top-4 right-4 z-10 flex items-center gap-2">
				{showLanguageSelect && onLanguageChange && (
					<Select value={language} onValueChange={onLanguageChange}>
						<SelectTrigger className="h-8 w-fit text-xs">
							<SelectValue />
						</SelectTrigger>
						<SelectContent>
							<SelectItem className="text-xs" value="python">
								Python
							</SelectItem>
							<SelectItem className="text-xs" value="typescript">
								TypeScript
							</SelectItem>
						</SelectContent>
					</Select>
				)}
				<Button variant="ghost" size="icon" onClick={copyToClipboard}>
					<Copy className="size-4" />
				</Button>
			</div>
			<CodeEditor className="w-full" code={code} lang={language} readonly={readonly} height={300} fontSize={14} options={EDITOR_OPTIONS} />
		</div>
	);
}

const CARDS = [
	{
		title: "What You'll See Here",
		description: "Real-time request logs from all your API calls",
		features: [
			"Real-time request logs from all your API calls",
			"Comprehensive request and error details",
			"Token usage, latency, and cost metrics",
			"Advanced filtering and search capabilities",
		],
	},
	{
		title: "Getting Started",
		description: "Use the examples below to get started",
		features: [
			"Choose an example from below",
			"Set Bifrost as your API endpoint",
			"Send a test request",
			"Monitor the response in real-time",
		],
	},
];

interface EmptyStateProps {
	isSocketConnected: boolean;
	error: string | null;
}

export function EmptyState({ isSocketConnected, error }: EmptyStateProps) {
	const [language, setLanguage] = useState<Language>("python");

	// Generate examples dynamically using the port utility
	const examples: Examples = useMemo(() => {
		const baseUrl = getExampleBaseUrl();

		return {
			curl: `curl -X POST ${baseUrl}/v1/chat/completions \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "openai/gpt-4o-mini",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ]
  }'`,
			sdk: {
				openai: {
					python: `import openai

client = openai.OpenAI(
    base_url="${baseUrl}/openai",
    api_key="dummy-api-key" # Handled by Bifrost
)

response = client.chat.completions.create(
    model="gpt-4o-mini", # or "provider/model" for other providers (anthropic/claude-3-sonnet)
    messages=[{"role": "user", "content": "Hello!"}]
)`,
					typescript: `import OpenAI from "openai";

const openai = new OpenAI({
  baseURL: "${baseUrl}/openai",
  apiKey: "dummy-api-key", // Handled by Bifrost
});

const response = await openai.chat.completions.create({
  model: "gpt-4o-mini", // or "provider/model" for other providers (anthropic/claude-3-sonnet)
  messages: [{ role: "user", content: "Hello!" }],
});`,
				},
				anthropic: {
					python: `import anthropic

client = anthropic.Anthropic(
    base_url="${baseUrl}/anthropic",
    api_key="dummy-api-key" # Handled by Bifrost
)

response = client.messages.create(
    model="claude-3-sonnet-20240229", # or "provider/model" for other providers (openai/gpt-4o-mini)
    max_tokens=1000,
    messages=[{"role": "user", "content": "Hello!"}]
)`,
					typescript: `import Anthropic from "@anthropic-ai/sdk";

const anthropic = new Anthropic({
  baseURL: "${baseUrl}/anthropic",
  apiKey: "dummy-api-key", // Handled by Bifrost
});

const response = await anthropic.messages.create({
  model: "claude-3-sonnet-20240229", // or "provider/model" for other providers (openai/gpt-4o-mini)
  max_tokens: 1000,
  messages: [{ role: "user", content: "Hello!" }],
});`,
				},
				genai: {
					python: `from google import genai
from google.genai.types import HttpOptions

client = genai.Client(
    api_key="dummy-api-key", # Handled by Bifrost
    http_options=HttpOptions(base_url="${baseUrl}/genai")
)

response = client.models.generate_content(
    model="gemini-2.5-pro", # or "provider/model" for other providers (openai/gpt-4o-mini)
    contents="Hello!"
)`,
					typescript: `import { GoogleGenerativeAI } from "@google/generative-ai";

const genAI = new GoogleGenerativeAI("dummy-api-key", { // Handled by Bifrost
  baseUrl: "${baseUrl}/genai",
});

const model = genAI.getGenerativeModel({ model: "gemini-2.5-pro" }); // or "provider/model" for other providers (openai/gpt-4o-mini)
const response = await model.generateContent("Hello!");`,
				},
				litellm: {
					python: `import litellm

litellm.api_base = "${baseUrl}/litellm"

response = litellm.completion(
    model="openai/gpt-4o-mini",
    messages=[{"role": "user", "content": "Hello!"}]
)`,
					typescript: `import { completion } from "litellm";

const response = await completion({
  model: "openai/gpt-4o-mini",
  messages: [{ role: "user", content: "Hello!" }],
  api_base: "${baseUrl}/litellm",
});`,
				},
			},
		};
	}, []);

	return (
		<div className="flex w-full flex-col items-center justify-center space-y-8">
			<div className="space-y-2 text-center">
				<h2 className="text-3xl font-bold">Welcome to Request Logs</h2>
				<p className="text-muted-foreground text-lg">Monitor and analyze all your API requests in real-time</p>
			</div>

			{isSocketConnected && (
				<div className="flex items-center justify-center gap-2 text-sm">
					<RefreshCw className="h-4 w-4 animate-spin" />
					<span>Listening for logs...</span>
				</div>
			)}

			{error && (
				<Alert>
					<AlertTriangle className="h-4 w-4" />
					<AlertDescription>{error}</AlertDescription>
				</Alert>
			)}

			<div className="grid w-full grid-cols-1 gap-6 md:grid-cols-2">
				{CARDS.map((card) => (
					<Card key={card.title} className="p-6">
						<h3 className="text-lg font-semibold">{card.title}</h3>
						<p className="text-muted-foreground">{card.description}</p>
						<ul className="text-muted-foreground space-y-3">
							{card.features.map((feature) => (
								<li key={feature} className="flex items-start gap-2">
									<ArrowRight className="mt-0.5 h-5 w-5 shrink-0" />
									<span>{feature}</span>
								</li>
							))}
						</ul>
					</Card>
				))}
			</div>

			<div className="w-full space-y-6">
				<div className="">
					<h3 className="text-xl font-semibold">Integration Examples</h3>
					<p className="text-muted-foreground">Send your first request to get started</p>
				</div>

				<Tabs defaultValue="curl" className="w-full">
					<TabsList className="grid h-10 w-full grid-cols-5">
						<TabsTrigger value="curl">cURL</TabsTrigger>
						<TabsTrigger value="openai">OpenAI SDK</TabsTrigger>
						<TabsTrigger value="anthropic">Anthropic SDK</TabsTrigger>
						<TabsTrigger value="genai">Google GenAI SDK</TabsTrigger>
						<TabsTrigger value="litellm">LiteLLM SDK</TabsTrigger>
					</TabsList>

					<TabsContent value="curl" className="mt-4">
						<CodeBlock code={examples.curl} language="bash" readonly={false} />
					</TabsContent>

					<TabsContent value="openai" className="mt-4">
						<CodeBlock
							code={examples.sdk.openai[language]}
							language={language}
							onLanguageChange={(newLang) => setLanguage(newLang as Language)}
							showLanguageSelect
						/>
					</TabsContent>

					<TabsContent value="anthropic" className="mt-4">
						<CodeBlock
							code={examples.sdk.anthropic[language]}
							language={language}
							onLanguageChange={(newLang) => setLanguage(newLang as Language)}
							showLanguageSelect
						/>
					</TabsContent>

					<TabsContent value="genai" className="mt-4">
						<CodeBlock
							code={examples.sdk.genai[language]}
							language={language}
							onLanguageChange={(newLang) => setLanguage(newLang as Language)}
							showLanguageSelect
						/>
					</TabsContent>

					<TabsContent value="litellm" className="mt-4">
						<CodeBlock
							code={examples.sdk.litellm[language]}
							language={language}
							onLanguageChange={(newLang) => setLanguage(newLang as Language)}
							showLanguageSelect
						/>
					</TabsContent>
				</Tabs>
			</div>
		</div>
	);
}
