"use client";
import { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { Deployment } from '@/types';

interface WebSocketContextType {
  deployments: Deployment[];
  updateDeployment: (deployment: Deployment) => void;
}

const WebSocketContext = createContext<WebSocketContextType | undefined>(undefined);

export function WebSocketProvider({ children }: { children: ReactNode }) {
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [ws, setWs] = useState<WebSocket | null>(null);

  useEffect(() => {
    const socket = new WebSocket('ws://localhost:8080/ws');
    setWs(socket);

    socket.onmessage = (event) => {
      const message = JSON.parse(event.data);
      if (message.type === 'status_update') {
        updateDeployment(message.data);
      }
    };

    socket.onclose = () => {
      // Attempt to reconnect after 5 seconds
      setTimeout(() => {
        setWs(null);
      }, 5000);
    };

    return () => {
      socket.close();
    };
  }, []);

  const updateDeployment = (updatedDeployment: Deployment) => {
    setDeployments((prevDeployments) => {
      const index = prevDeployments.findIndex(d => d.name === updatedDeployment.name);
      if (index === -1) {
        return [...prevDeployments, updatedDeployment];
      }
      const newDeployments = [...prevDeployments];
      newDeployments[index] = updatedDeployment;
      return newDeployments;
    });
  };

  return (
    <WebSocketContext.Provider value={{ deployments, updateDeployment }}>
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