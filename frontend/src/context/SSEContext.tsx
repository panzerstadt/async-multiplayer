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

    newEventSource.addEventListener("new_save", ({ data }) => {
      const parsed = JSON.parse(data);
      toast.info(`New save for game ${parsed.game_id}: ${parsed.message}`, {
        duration: Number.POSITIVE_INFINITY,
        dismissible: true,
        cancel: { label: "ok!", onClick: () => {} },
      });
    });

    newEventSource.addEventListener("broadcast", ({ data }) => {
      const parsed = JSON.parse(data);
      toast(() => (
        <pre className="text-blue-500 whitespace-pre-wrap">{parsed.replace("\\n", "\n")}</pre>
      ));
    });

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
