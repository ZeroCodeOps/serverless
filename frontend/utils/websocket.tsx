"use client";
import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { Deployment } from '../types';
import { showErrorAlert, showInfoAlert, showSuccessAlert, showWarningAlert, showConfirmDialog } from './alert';

interface WebSocketContextType {
  deployments: Deployment[];
  updateDeployment: (deployment: Deployment) => void;
  ws: WebSocket | null;
  isConnected: boolean;
  showAlert: (type: 'success' | 'error' | 'info' | 'warning', message: string) => void;
  showConfirmDialog: (title: string, text: string) => Promise<boolean>;
}

const WebSocketContext = createContext<WebSocketContextType | null>(null);

export const useWebSocket = () => {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error('useWebSocket must be used within a WebSocketProvider');
  }
  return context;
};

interface WebSocketProviderProps {
  children: ReactNode;
}

export const WebSocketProvider: React.FC<WebSocketProviderProps> = ({ children }) => {
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [ws, setWs] = useState<WebSocket | null>(null);
  const [isConnected, setIsConnected] = useState(false);

  useEffect(() => {
    const socket = new WebSocket('ws://localhost:8080/ws');
    
    socket.onopen = () => {
      setIsConnected(true);
      showSuccessAlert('Connected to server');
    };

    socket.onclose = () => {
      setIsConnected(false);
      showErrorAlert('Disconnected from server');
    };

    socket.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === 'status_update') {
        updateDeployment(data.data);
      }
    };

    setWs(socket);

    return () => {
      socket.close();
    };
  }, []);

  const updateDeployment = (deployment: Deployment) => {
    setDeployments((prev) => {
      const index = prev.findIndex((d) => d.id === deployment.id);
      if (index === -1) {
        return [...prev, deployment];
      }
      const newDeployments = [...prev];
      newDeployments[index] = deployment;
      return newDeployments;
    });
  };

  const showAlert = (type: 'success' | 'error' | 'info' | 'warning', message: string) => {
    switch (type) {
      case 'success':
        showSuccessAlert(message);
        break;
      case 'error':
        showErrorAlert(message);
        break;
      case 'info':
        showInfoAlert(message);
        break;
      case 'warning':
        showWarningAlert(message);
        break;
    }
  };

  const value: WebSocketContextType = {
    deployments,
    updateDeployment,
    ws,
    isConnected,
    showAlert,
    showConfirmDialog,
  };

  return (
    <WebSocketContext.Provider value={value}>
      {children}
    </WebSocketContext.Provider>
  );
}; 