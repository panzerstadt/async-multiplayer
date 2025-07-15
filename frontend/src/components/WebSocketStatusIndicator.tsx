"use client";

import { useWebSocket } from "@/context/WebSocketContext";
import { useEffect, useState } from "react";

export default function WebSocketStatusIndicator() {
  const { socket } = useWebSocket();
  const [isConnected, setIsConnected] = useState(false);

  useEffect(() => {
    if (!socket) return;

    setIsConnected(socket.connected);

    socket.on("connect", () => {
      setIsConnected(true);
    });

    socket.on("disconnect", () => {
      setIsConnected(false);
    });

    return () => {
      socket.off("connect");
      socket.off("disconnect");
    };
  }, [socket]);

  return (
    <div
      className={`fixed bottom-4 right-4 px-3 py-1 rounded-full text-sm font-semibold
        ${isConnected ? "bg-green-500 text-white" : "bg-red-500 text-white"}
      `}
    >
      WebSocket: {isConnected ? "Connected" : "Disconnected"}
    </div>
  );
}
