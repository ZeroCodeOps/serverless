"use client";

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Header from '@/components/Header';
import DeploymentTable from '@/components/DeploymentTable';
import { NewDeploymentDialog } from '@/components/NewDeploymentDialog';
import { useAuth } from '@/utils/auth';
import { mockDeployments } from '@/utils/mockData';
import { Deployment } from '@/types';
import { NextPage } from 'next';
import { BACKEND_URL } from '@/lib/utils';

const Dashboard: NextPage = () => {
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isNewDeploymentDialogOpen, setIsNewDeploymentDialogOpen] = useState<boolean>(false);
  const isAuthenticated = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (isAuthenticated) {
      const fetchDeployments = async () => {
        const response = await fetch(`${BACKEND_URL}/deployments/`);
        const data = await response.json();
        setDeployments(data);
      };
      // Simulate API call
      setTimeout(() => {
        fetchDeployments();
        setIsLoading(false);
      }, 500);
    }
  }, [isAuthenticated]);

  const handleOpenNewDeploymentDialog = (): void => {
    setIsNewDeploymentDialogOpen(true);
  };

  const handleCreateDeployment = (name: string, language: 'node' | 'go' | 'python'): void => {
    const newId = String(deployments.length + 1);
    const newDeployment: Deployment = {
      id: newId,
      name: name,
      status: 'Stopped',
      createdAt: new Date().toISOString().split('T')[0],
      language: language,
    };
    
    setDeployments([...deployments, newDeployment]);
    setIsNewDeploymentDialogOpen(false);
    
    // Optionally, redirect to edit page for the new deployment
    router.push(`/edit/${newId}`);
  };

  const handleDeleteDeployment = (id: string): void => {
    setDeployments(deployments.filter(deployment => deployment.id !== id));
  };
  
  const handleToggleDeployment = (id: string): void => {
    setDeployments(
      deployments.map((d) => {
        if (d.id === id) {
          return {
            ...d,
            status: d.status === 'Running' ? 'Stopped' : 'Running',
          };
        }
        return d;
      })
    );
  };

  if (!isAuthenticated) {
    return <div className="min-h-screen flex items-center justify-center">Loading...</div>;
  }

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Header />
      <main className="container mx-auto py-8 px-4 flex-1">
        <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 mb-8">
          <div>
            <h1 className="text-2xl font-bold">Your Deployments</h1>
            <p className="text-muted-foreground">Manage your serverless functions</p>
          </div>
          <button
            onClick={handleOpenNewDeploymentDialog}
            className="btn btn-primary"
          >
            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            Create New Deployment
          </button>
        </div>
        
        {isLoading ? (
          <div className="flex flex-col items-center justify-center py-12">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
            <p className="mt-4 text-muted-foreground">Loading deployments...</p>
          </div>
        ) : (
          <DeploymentTable 
            deployments={deployments} 
            onDelete={handleDeleteDeployment} 
            onToggle={handleToggleDeployment}
          />
        )}
      </main>
      
      <NewDeploymentDialog
        isOpen={isNewDeploymentDialogOpen}
        onConfirm={handleCreateDeployment}
        onCancel={() => setIsNewDeploymentDialogOpen(false)}
      />
    </div>
  );
};

export default Dashboard;