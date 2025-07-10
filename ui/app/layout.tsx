import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import Sidebar from "@/components/sidebar";
import { SidebarProvider } from "@/components/ui/sidebar";
import { ThemeProvider } from "@/components/theme-provider";
import { Toaster } from "sonner";
import ProgressProvider from "@/components/progress-bar";
import { WebSocketProvider } from "@/hooks/useWebSocket";

const geistSans = Geist({
	variable: "--font-geist-sans",
	subsets: ["latin"],
});

const geistMono = Geist_Mono({
	variable: "--font-geist-mono",
	subsets: ["latin"],
});

export const metadata: Metadata = {
	title: "Bifrost - The fastest LLM gateway",
	description:
		"Production-ready fastest LLM gateway that connects to 8+ providers through a single API. Get automatic failover, load balancing, mcp support and zero-downtime deployments.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
	return (
		<html lang="en" suppressHydrationWarning>
			<body className={`${geistSans.variable} ${geistMono.variable} antialiased`}>
				<ProgressProvider>
					<ThemeProvider attribute="class" defaultTheme="system" enableSystem>
						<Toaster />
						<WebSocketProvider>
							<SidebarProvider>
								<Sidebar />
								<main className="custom-scrollbar relative mx-auto flex min-h-screen w-5xl flex-col py-12">{children}</main>
							</SidebarProvider>
						</WebSocketProvider>
					</ThemeProvider>
				</ProgressProvider>
			</body>
		</html>
	);
}
