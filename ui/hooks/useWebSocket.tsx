"use client";

import React, { createContext, useContext, useEffect, useRef, useState, type ReactNode } from "react";
import type { LogEntry } from "../lib/types/logs";

interface WebSocketContextType {
	isConnected: boolean;
	ws: React.RefObject<WebSocket | null>;
	setMessageHandler: (handler: (log: LogEntry) => void) => void;
}

const WebSocketContext = createContext<WebSocketContextType | null>(null);

declare const process: {
	env: {
		NEXT_PUBLIC_BIFROST_PORT?: string;
	};
};

interface WebSocketProviderProps {
	children: ReactNode;
	onMessage?: (log: LogEntry) => void;
}

// Global reference to maintain state across component remounts
let globalWsRef: WebSocket | null = null;
let globalMessageHandler: ((log: LogEntry) => void) | null = null;

export function WebSocketProvider({ children, onMessage }: WebSocketProviderProps) {
	const wsRef = useRef<WebSocket | null>(globalWsRef);
	const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
	const [isConnected, setIsConnected] = useState(false);

	const setMessageHandler = (handler: (log: LogEntry) => void) => {
		globalMessageHandler = handler;
	};

	useEffect(() => {
		const port = process.env.NEXT_PUBLIC_BIFROST_PORT || "8080";
		const connect = () => {
			if (wsRef.current?.readyState === WebSocket.OPEN) {
				return;
			}

			const ws = new WebSocket(`ws://localhost:${port}/ws/logs`);
			wsRef.current = ws;
			globalWsRef = ws;

			ws.onopen = () => {
				console.log("WebSocket connected");
				setIsConnected(true);
				// Clear any pending reconnection attempts
				if (reconnectTimeoutRef.current) {
					clearTimeout(reconnectTimeoutRef.current);
					reconnectTimeoutRef.current = null;
				}
			};

			ws.onmessage = (event) => {
				try {
					const data = JSON.parse(event.data);
					if (data.type === "log") {
						if (globalMessageHandler) {
							globalMessageHandler(data.payload);
						} else if (onMessage) {
							onMessage(data.payload);
						}
					}
				} catch (error) {
					console.error("Failed to parse WebSocket message:", error);
				}
			};

			ws.onclose = () => {
				console.log("WebSocket disconnected, attempting to reconnect...");
				setIsConnected(false);
				// Attempt to reconnect after 5 seconds
				reconnectTimeoutRef.current = setTimeout(connect, 5000);
			};

			ws.onerror = (error) => {
				setIsConnected(false);
				ws.close();
			};
		};

		connect();

		// Cleanup function
		return () => {
			// Don't close the WebSocket on unmount since it's global
			if (reconnectTimeoutRef.current) {
				clearTimeout(reconnectTimeoutRef.current);
				reconnectTimeoutRef.current = null;
			}
		};
	}, [onMessage]);

	return <WebSocketContext.Provider value={{ isConnected, ws: wsRef, setMessageHandler }}>{children}</WebSocketContext.Provider>;
}

export function useWebSocket() {
	const context = useContext(WebSocketContext);
	if (!context) {
		throw new Error("useWebSocket must be used within a WebSocketProvider");
	}
	return context;
}
