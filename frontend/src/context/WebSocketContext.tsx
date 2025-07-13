"use client";

import { createContext, useContext, useEffect, ReactNode, useState } from 'react';
import { io as ClientIO, Socket } from 'socket.io-client';
import { toast } from 'sonner';

interface WebSocketContextType {
  socket: Socket | null;
}

const WebSocketContext = createContext<WebSocketContextType | undefined>(undefined);

export function WebSocketProvider({ children }: { children: ReactNode }) {
  const [socket, setSocket] = useState<Socket | null>(null);

  useEffect(() => {
    const newSocket = ClientIO(process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080');
    setSocket(newSocket);

    newSocket.on('connect', () => {
      console.log('Connected to WebSocket server');
    });

    newSocket.on('disconnect', () => {
      console.log('Disconnected from WebSocket server');
    });

    newSocket.on('new_save', (data: { game_id: string; message: string }) => {
      toast.info(`New save for game ${data.game_id}: ${data.message}`);
    });

    return () => {
      newSocket.disconnect();
    };
  }, []);

  return (
    <WebSocketContext.Provider value={{ socket }}>
      {children}
    </WebSocketContext.Provider>
  );
}

export function useWebSocket() {
  const context = useContext(WebSocketContext);
  if (context === undefined) {
    throw new Error('useWebSocket must be used within a WebSocketProvider');
  }
  return context;
}
