"use client";

import { Home, BookOpen, Settings, Puzzle, Zap, ExternalLink, ChevronRight } from "lucide-react";

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
import { Icons } from "@/lib/constants/icons";

// Main navigation items
const navigationItems = [
	{
		title: "Logs",
		url: "/",
		icon: Home,
		description: "Request logs & monitoring",
		badge: "Live",
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
		title: "GitHub Repository",
		url: "https://github.com/maximhq/bifrost",
		icon: ExternalLink,
	},
	{
		title: "Full Documentation",
		url: "https://github.com/maximhq/bifrost/tree/main/docs",
		icon: BookOpen,
	},
];

export default function AppSidebar() {
	const pathname = usePathname();

	const isActiveRoute = (url: string) => {
		if (url === "/" && pathname === "/") return true;
		if (url !== "/" && pathname.startsWith(url)) return true;
		return false;
	};

	return (
		<Sidebar className="border-border border-r">
			<SidebarHeader className="flex h-12 justify-center">
				<Link href="/" className="group flex items-center gap-2">
					<div className="from-primary flex h-6 w-6 items-center justify-center">{Icons.bifrost}</div>
					<h1 className="text-foreground text-lg leading-none font-bold">Bifrost</h1>
				</Link>
			</SidebarHeader>

			<SidebarSeparator />

			<SidebarContent>
				<SidebarGroup>
					<SidebarGroupLabel className="text-muted-foreground px-2 py-2 text-xs font-semibold tracking-wider uppercase">
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
												<div className="flex items-center space-x-2">
													{item.badge && (
														<Badge
															variant={item.badge === "Live" ? "default" : "secondary"}
															className={cn("h-5 px-2 py-0.5 text-xs", item.badge === "Live" && "animate-pulse duration-200")}
														>
															{item.badge}
														</Badge>
													)}
													{isActive && <ChevronRight className="text-primary h-3 w-3" />}
												</div>
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
					<SidebarGroupLabel className="text-muted-foreground px-2 py-2 text-xs font-semibold tracking-wider uppercase">
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
												<item.icon className="text-muted-foreground group-hover:text-foreground h-4 w-4" />
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
