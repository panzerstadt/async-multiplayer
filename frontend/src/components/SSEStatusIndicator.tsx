"use client";

import { useSSE } from "@/context/SSEContext";
import { useEffect, useState } from "react";

export default function SSEStatusIndicator() {
  const { eventSource } = useSSE();
  const [isConnected, setIsConnected] = useState(false);

  useEffect(() => {
    if (!eventSource) return;

    // EventSource.OPEN (0) means connection is established
    setIsConnected(eventSource.readyState === EventSource.OPEN);

    eventSource.onopen = () => {
      setIsConnected(true);
    };

    eventSource.onerror = () => {
      setIsConnected(false);
    };

    return () => {
      eventSource.onopen = null;
      eventSource.onerror = null;
    };
  }, [eventSource]);

  return (
    <div
      className={`fixed bottom-4 right-4 px-3 py-1 rounded-full text-sm font-semibold
        ${isConnected ? "bg-green-500 text-white" : "bg-red-500 text-white"}
      `}
    >
      SSE: {isConnected ? "Connected" : "Disconnected"}
    </div>
  );
}
