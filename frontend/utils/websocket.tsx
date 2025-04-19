"use client";
import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { Deployment } from '../types';
import { showErrorAlert, showInfoAlert, showSuccessAlert, showWarningAlert, showConfirmDialog } from './alert';

interface WebSocketContextType {
  deployments: Deployment[];
  setDeployments: (deployments: Deployment[]) => void;
  updateDeployment: (deployment: Deployment) => void;
  ws: WebSocket | null;
  isConnected: boolean;
  showAlert: (type: 'success' | 'error' | 'info' | 'warning', message: string) => void;
  showConfirmDialog: (title: string, text: string) => Promise<boolean>;
  loadingDeployments: Set<string>;
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
  const [loadingDeployments, setLoadingDeployments] = useState<Set<string>>(new Set());

  // Fetch initial deployments
  useEffect(() => {
    const fetchDeployments = async () => {
      try {
        const response = await fetch('http://localhost:8080/deployments');
        if (!response.ok) throw new Error('Failed to fetch deployments');
        const data = await response.json();
        setDeployments(data);
      } catch (error) {
        console.error('Error fetching deployments:', error);
        showErrorAlert('Failed to load deployments');
      }
    };

    fetchDeployments();
  }, []);

  useEffect(() => {
    const socket = new WebSocket('ws://localhost:8080/ws');

    const fetchDeployments = async () => {
      try {
        const response = await fetch('http://localhost:8080/deployments');
        if (!response.ok) throw new Error('Failed to fetch deployments');
        const data = await response.json();
        setDeployments(data);
      } catch (error) {
        console.error('Error fetching deployments:', error);
        showErrorAlert('Failed to load deployments');
      }
    };

    socket.onopen = () => {
      setIsConnected(true);
      fetchDeployments();
      showSuccessAlert('Connected to server');
    };

    socket.onclose = () => {
      setIsConnected(false);
      showErrorAlert('Disconnected from server');
    };

    socket.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === 'status_update') {
        const deployment = data.data as Deployment;
        updateDeployment(deployment);

        // Handle status-specific alerts
        switch (deployment.status) {
          case 'Starting':
            setLoadingDeployments(prev => new Set(prev).add(deployment.id));
            showInfoAlert(`Starting function ${deployment.name}...`);
            break;
          case 'Running':
            setLoadingDeployments(prev => {
              const newSet = new Set(prev);
              newSet.delete(deployment.id);
              return newSet;
            });
            showSuccessAlert(`Function ${deployment.name} is running on port ${deployment.port}`);
            break;
          case 'Failed':
            setLoadingDeployments(prev => {
              const newSet = new Set(prev);
              newSet.delete(deployment.id);
              return newSet;
            });
            showErrorAlert(`Function ${deployment.name} failed to start`);
            break;
          case 'Stopped':
            setLoadingDeployments(prev => {
              const newSet = new Set(prev);
              newSet.delete(deployment.id);
              return newSet;
            });
            showInfoAlert(`Function ${deployment.name} has stopped`);
            break;
        }
      }
    };

    setWs(socket);

    return () => {
      socket.close();
    };
  }, []);

  const updateDeployment = (updatedDeployment: Deployment) => {
    setDeployments(prevDeployments => {
      // Find the index of the deployment to update
      if (!prevDeployments) return [updatedDeployment];
      const index = prevDeployments.findIndex(d => d.id === updatedDeployment.id);
      
      if (index === -1) {
        // If deployment doesn't exist, add it to the list
        return [...prevDeployments, updatedDeployment];
      }
      
      // Create a new array with the updated deployment
      const newDeployments = [...prevDeployments];
      newDeployments[index] = updatedDeployment;
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
    setDeployments,
    updateDeployment,
    ws,
    isConnected,
    showAlert,
    showConfirmDialog,
    loadingDeployments,
  };

  return (
    <WebSocketContext.Provider value={value}>
      {children}
    </WebSocketContext.Provider>
  );
}; 