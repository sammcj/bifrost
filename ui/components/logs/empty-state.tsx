"use client";

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Copy, RefreshCw, ArrowRight, AlertTriangle } from "lucide-react";
import { CodeEditor } from "./ui/code-editor";
import { toast } from "sonner";
import { useState } from "react";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Alert, AlertDescription } from "../ui/alert";

type Provider = "openai" | "anthropic" | "genai";
type Language = "python" | "typescript";

type Examples = {
	curl: string;
	sdk: {
		[P in Provider]: {
			[L in Language]: string;
		};
	};
};

const EXAMPLES: Examples = {
	curl: `curl -X POST http://localhost:8080/v1/chat/completions \\
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
    base_url="http://localhost:8080/openai",
    api_key="dummy-api-key" # Handled by Bifrost
)

response = client.chat.completions.create(
    model="gpt-4o-mini",
    messages=[{"role": "user", "content": "Hello!"}]
)`,
			typescript: `import OpenAI from "openai";

const openai = new OpenAI({
  baseURL: "http://localhost:8080/openai",
  apiKey: "dummy-api-key", // Handled by Bifrost
});

const response = await openai.chat.completions.create({
  model: "gpt-4o-mini",
  messages: [{ role: "user", content: "Hello!" }],
});`,
		},
		anthropic: {
			python: `import anthropic

client = anthropic.Anthropic(
    base_url="http://localhost:8080/anthropic",
    api_key="dummy-api-key" # Handled by Bifrost
)

response = client.messages.create(
    model="claude-3-sonnet-20240229",
    max_tokens=1000,
    messages=[{"role": "user", "content": "Hello!"}]
)`,
			typescript: `import Anthropic from "@anthropic-ai/sdk";

const anthropic = new Anthropic({
  baseURL: "http://localhost:8080/anthropic",
  apiKey: "dummy-api-key", // Handled by Bifrost
});

const response = await anthropic.messages.create({
  model: "claude-3-sonnet-20240229",
  max_tokens: 1000,
  messages: [{ role: "user", content: "Hello!" }],
});`,
		},
		genai: {
			python: `from google import genai
from google.genai.types import HttpOptions

client = genai.Client(
    api_key="dummy-api-key", # Handled by Bifrost
    http_options=HttpOptions(base_url="http://localhost:8080/genai")
)

response = client.models.generate_content(
    model="gemini-pro",
    contents="Hello!"
)`,
			typescript: `import { GoogleGenerativeAI } from "@google/generative-ai";

const genAI = new GoogleGenerativeAI("dummy-api-key", { // Handled by Bifrost
  baseUrl: "http://localhost:8080/genai",
});

const model = genAI.getGenerativeModel({ model: "gemini-pro" });
const response = await model.generateContent("Hello!");`,
		},
	},
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

			<div className="w-full">
				<h3 className="mb-4 text-lg font-semibold">Integration Examples</h3>
				<Tabs defaultValue="curl" className="w-full">
					<TabsList className="h-10 w-full justify-start">
						<TabsTrigger value="curl">cURL</TabsTrigger>
						<TabsTrigger value="openai">OpenAI SDK</TabsTrigger>
						<TabsTrigger value="anthropic">Anthropic SDK</TabsTrigger>
						<TabsTrigger value="genai">Google GenAI SDK</TabsTrigger>
					</TabsList>

					{/* cURL Example Tab */}
					<TabsContent value="curl" className="p-4">
						<CodeBlock code={EXAMPLES.curl} language="bash" readonly={false} />
					</TabsContent>

					{/* SDK Tabs */}
					{(Object.keys(EXAMPLES.sdk) as Provider[]).map((provider) => (
						<TabsContent key={provider} value={provider} className="space-y-4 p-4">
							<CodeBlock
								code={EXAMPLES.sdk[provider][language]}
								language={language === "typescript" ? "typescript" : "python"}
								onLanguageChange={(lang) => setLanguage(lang as Language)}
								showLanguageSelect={true}
							/>
						</TabsContent>
					))}
				</Tabs>
			</div>
		</div>
	);
}
