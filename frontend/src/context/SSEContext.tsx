"use client";

import { createContext, useContext, useEffect, ReactNode, useState } from "react";
import { toast } from "sonner";

interface SSEContextType {
  eventSource: EventSource | null;
}

const SSEContext = createContext<SSEContextType | undefined>(undefined);

export function SSEProvider({ children }: { children: ReactNode }) {
  const [eventSource, setEventSource] = useState<EventSource | null>(null);

  useEffect(() => {
    console.log("Attempting to connect to SSE endpoint...");
    const newEventSource = new EventSource(`${process.env.NEXT_PUBLIC_API_URL}/sse/notifications`);
    setEventSource(newEventSource);

    newEventSource.onopen = () => {
      console.log("SSE: Connected");
    };

    newEventSource.onmessage = (event) => {
      console.log("SSE: Message received", event.data);
      try {
        const data = JSON.parse(event.data);
        if (data.event === "new_save") {
          toast.info(`New save for game ${data.game_id}: ${data.message}`, {
            duration: Number.POSITIVE_INFINITY,
            dismissible: true,
          });
        }
      } catch (e) {
        console.error("SSE: Error parsing message", e);
      }
    };

    newEventSource.onerror = (error) => {
      console.error("SSE: Connection Error", error);
      newEventSource.close();
    };

    return () => {
      newEventSource.close();
    };
  }, []);

  return <SSEContext.Provider value={{ eventSource }}>{children}</SSEContext.Provider>;
}

export function useSSE() {
  const context = useContext(SSEContext);
  if (context === undefined) {
    throw new Error("useSSE must be used within an SSEProvider");
  }
  return context;
}
