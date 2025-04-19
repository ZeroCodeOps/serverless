"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import Header from "@/components/Header";
import DeploymentTable from "@/components/DeploymentTable";
import { NewDeploymentDialog } from "@/components/NewDeploymentDialog";
import { useAuth } from "@/utils/auth";
import { Deployment } from "@/types";
import { NextPage } from "next";
import { BACKEND_URL } from "@/lib/utils";
import { useWebSocket } from "@/utils/websocket";

const Dashboard: NextPage = () => {
  const { deployments: wsDeployments } = useWebSocket();
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [refresh, setRefresh] = useState<boolean>(false);
  const [isNewDeploymentDialogOpen, setIsNewDeploymentDialogOpen] =
    useState<boolean>(false);
  const isAuthenticated = useAuth();

  useEffect(() => {
    if (isAuthenticated) {
      const fetchDeployments = async () => {
        try {
          const response = await fetch(`${BACKEND_URL}/deployments/`);
          if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
          }
          const data = await response.json();
          setDeployments(data);
        } catch (error) {
          console.error("Error fetching deployments:", error);
        } finally {
          setIsLoading(false);
        }
      };
      fetchDeployments();
    }
  }, [isAuthenticated, refresh]);

  // Update deployments when WebSocket updates arrive
  useEffect(() => {
    if (wsDeployments.length > 0) {
      setDeployments(wsDeployments);
    }
  }, [wsDeployments]);

  const handleOpenNewDeploymentDialog = (): void => {
    setIsNewDeploymentDialogOpen(true);
  };

  const handleCreateDeployment = async (
    name: string,
    language: "node" | "go" | "python",
  ) => {
    try {
      const response = await fetch(`${BACKEND_URL}/create/${language}`, {
        method: "POST",
        body: new URLSearchParams({
          name,
        }),
        headers: {
          "Content-Type": "application/x-www-form-urlencoded",
        },
      });
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || response.statusText);
      }
      setRefresh(!refresh);
    } catch (error) {
      alert(error instanceof Error ? error.message : "Failed to create deployment");
    } finally {
      setIsNewDeploymentDialogOpen(false);
    }
  };

  const handleDeleteDeployment = async (name: string) => {
    try {
      const response = await fetch(`${BACKEND_URL}/deployments/${name}`, {
        method: "DELETE",
      });
      if (!response.ok) {
        throw new Error(`Failed to delete deployment: ${response.statusText}`);
      }
      setRefresh(!refresh);
    } catch (error) {
      alert(error instanceof Error ? error.message : "Failed to delete deployment");
    }
  };
  
  const handleBuildDeployment = async (name: string) => {
    try {
      const response = await fetch(`${BACKEND_URL}/build/${name}`, {
        method: "POST",
      });
      if (!response.ok) {
        throw new Error(`Failed to build deployment: ${response.statusText}`);
      }
      setRefresh(!refresh);
    } catch (error) {
      alert(error instanceof Error ? error.message : "Failed to build deployment");
    }
  };

  const handleToggleDeployment = async (name: string) => {
    try {
      const deployment = deployments.find(d => d.name === name);
      if (!deployment) return;

      const endpoint = deployment.status === "Running" ? "stop" : "start";
      const response = await fetch(`${BACKEND_URL}/${endpoint}/${name}`, {
        method: "POST",
      });
      if (!response.ok) {
        throw new Error(`Failed to ${endpoint} deployment: ${response.statusText}`);
      }
      setRefresh(!refresh);
    } catch (error) {
      alert(error instanceof Error ? error.message : "Failed to toggle deployment");
    }
  };

  if (!isAuthenticated) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        Loading...
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Header />
      <main className="container mx-auto py-8 px-4 flex-1">
        <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 mb-8">
          <div>
            <h1 className="text-2xl font-bold">Your Deployments</h1>
            <p className="text-muted-foreground">
              Manage your serverless functions
            </p>
          </div>
          <button
            onClick={handleOpenNewDeploymentDialog}
            className="btn btn-primary"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-5 w-5 mr-2"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 4v16m8-8H4"
              />
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
            onBuild={handleBuildDeployment}
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
