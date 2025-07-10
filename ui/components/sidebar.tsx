"use client";

import { Home, BookOpen, Settings, Puzzle, ExternalLink, HeartHandshake } from "lucide-react";

import {
	Sidebar,
	SidebarContent,
	SidebarGroup,
	SidebarGroupContent,
	SidebarGroupLabel,
	SidebarHeader,
	SidebarMenu,
	SidebarMenuButton,
	SidebarMenuItem,
	SidebarSeparator,
	SidebarFooter,
} from "@/components/ui/sidebar";
import { Badge } from "@/components/ui/badge";
import { usePathname } from "next/navigation";
import Link from "next/link";
import { cn } from "@/lib/utils";
import { useTheme } from "next-themes";
import { useState, useEffect } from "react";
import Image from "next/image";
import { ThemeToggle } from "./theme-toggle";
import { useWebSocket } from "@/hooks/useWebSocket";
import { BookOpenTextIcon, DiscordLogoIcon, GithubLogoIcon } from "@phosphor-icons/react";

// Main navigation items
const navigationItems = [
	{
		title: "Logs",
		url: "/",
		icon: Home,
		description: "Request logs & monitoring",
	},
	{
		title: "Config",
		url: "/config",
		icon: Settings,
		description: "Providers & MCP configuration",
	},
	{
		title: "Docs",
		url: "/docs",
		icon: BookOpen,
		description: "Documentation & guides",
	},
	{
		title: "Plugins",
		url: "/plugins",
		icon: Puzzle,
		description: "Extend Bifrost functionality",
		badge: "Beta",
	},
];

// External links
const externalLinks = [
	{
		title: "Discord Server",
		url: "https://getmax.im/bifrost-discord",
		icon: DiscordLogoIcon,
	},
	{
		title: "GitHub Repository",
		url: "https://github.com/maximhq/bifrost",
		icon: GithubLogoIcon,
	},
	{
		title: "Full Documentation",
		url: "https://github.com/maximhq/bifrost/tree/main/docs",
		icon: BookOpenTextIcon,
	},
];

export default function AppSidebar() {
	const pathname = usePathname();
	const [mounted, setMounted] = useState(false);
	const { resolvedTheme } = useTheme();

	useEffect(() => {
		setMounted(true);
	}, []);

	const isActiveRoute = (url: string) => {
		if (url === "/" && pathname === "/") return true;
		if (url !== "/" && pathname.startsWith(url)) return true;
		return false;
	};

	// Always render the light theme version for SSR to avoid hydration mismatch
	const logoSrc = mounted && resolvedTheme === "dark" ? "/bifrost-logo-dark.png" : "/bifrost-logo.png";

	const { isConnected: isWebSocketConnected } = useWebSocket();

	return (
		<Sidebar className="border-border border-r">
			<SidebarHeader className="flex h-12 justify-between border-b px-0">
				<div className="flex h-full items-center justify-between gap-2 px-1.5">
					<Link href="/" className="group flex items-center gap-2">
						<Image className="h-10 w-auto" src={logoSrc} alt="Bifrost" width={100} height={100} />
					</Link>
					<ThemeToggle />
				</div>
			</SidebarHeader>

			<SidebarContent>
				<SidebarGroup>
					<SidebarGroupLabel className="text-muted-foreground px-3 py-2 text-xs font-semibold tracking-wider uppercase">
						Navigation
					</SidebarGroupLabel>
					<SidebarGroupContent>
						<SidebarMenu className="space-y-1">
							{navigationItems.map((item) => {
								const isActive = isActiveRoute(item.url);
								return (
									<SidebarMenuItem key={item.title}>
										<SidebarMenuButton
											asChild
											className={`relative h-16 rounded-lg border px-3 transition-all duration-200 ${
												isActive
													? "bg-accent text-primary border-primary/20 shadow-sm"
													: "hover:bg-accent hover:text-accent-foreground border-transparent"
											} `}
										>
											<Link href={item.url} className="flex w-full items-center justify-between">
												<div>
													<div className="flex items-center gap-2">
														<item.icon className={`h-4 w-4 ${isActive ? "text-primary" : "text-muted-foreground"}`} />
														<span className="text-sm font-medium">{item.title}</span>
													</div>
													<span className="text-muted-foreground overflow-hidden text-xs text-ellipsis whitespace-nowrap">
														{item.description}
													</span>
												</div>
												{item.url === "/" && isWebSocketConnected && (
													<div className="h-2 w-2 animate-pulse rounded-full bg-green-800 dark:bg-green-200" />
												)}
												{item.badge && (
													<Badge
														variant={item.badge === "Live" ? "default" : "secondary"}
														className={cn("h-5 px-2 py-0.5 text-xs", item.badge === "Live" && "animate-pulse duration-200")}
													>
														{item.badge}
													</Badge>
												)}
											</Link>
										</SidebarMenuButton>
									</SidebarMenuItem>
								);
							})}
						</SidebarMenu>
					</SidebarGroupContent>
				</SidebarGroup>

				<SidebarSeparator className="my-4" />

				<SidebarGroup>
					<SidebarGroupLabel className="text-muted-foreground px-3 py-2 text-xs font-semibold tracking-wider uppercase">
						Resources
					</SidebarGroupLabel>
					<SidebarGroupContent>
						<SidebarMenu className="space-y-1">
							{externalLinks.map((item) => (
								<SidebarMenuItem key={item.title}>
									<SidebarMenuButton
										asChild
										className="hover:bg-accent hover:text-accent-foreground h-9 rounded-lg px-3 transition-all duration-200"
									>
										<a href={item.url} target="_blank" rel="noopener noreferrer" className="group flex w-full items-center justify-between">
											<div className="flex items-center space-x-3">
												<item.icon className="text-muted-foreground h-4 w-4" size={16} weight="bold" />
												<span className="text-sm">{item.title}</span>
											</div>
											<ExternalLink className="text-muted-foreground h-3 w-3 opacity-0 transition-opacity group-hover:opacity-100" />
										</a>
									</SidebarMenuButton>
								</SidebarMenuItem>
							))}
						</SidebarMenu>
					</SidebarGroupContent>
				</SidebarGroup>
			</SidebarContent>

			<SidebarFooter className="border-border border-t px-6 py-4">
				<div className="text-muted-foreground mx-auto flex w-fit items-center space-x-1 text-xs">
					<span>Made with ♥️ by</span>
					<a href="https://getmaxim.ai" target="_blank" rel="noopener noreferrer" className="text-primary">
						Maxim AI
					</a>
				</div>
			</SidebarFooter>
		</Sidebar>
	);
}
