"use client";

import { useState, useEffect } from "react";
import Header from "@/components/Header";
import CodeEditor from "@/components/Editor";
import { useAuth } from "@/utils/auth";
import { mockFiles } from "@/utils/mockData";
import { NextPage } from "next";

const EditDeployment: NextPage = () => {
  const [id, setId] = useState<string>("");
  const [packageFile, setPackageFile] = useState<string>("");
  const [codeFile, setCodeFile] = useState<string>("");
  const [loading, setLoading] = useState<boolean>(true);
  const [isSaving, setIsSaving] = useState<boolean>(false);
  const [saveSuccess, setSaveSuccess] = useState<boolean>(false);
  const isAuthenticated = useAuth();
  
  useEffect(() => {
    setId(window.location.pathname.split("/").pop() || "");
  }, []);

  useEffect(() => {
    if (isAuthenticated && id && typeof id === "string") {
      // Simulate API call
      setTimeout(() => {
        if (mockFiles[id]) {
          setPackageFile(mockFiles[id].packageFile);
          setCodeFile(mockFiles[id].codeFile);
        }
        setLoading(false);
      }, 600);
    }
  }, [isAuthenticated, id]);

  const handleSave = (): void => {
    // Simulate API call
    setIsSaving(true);
    setSaveSuccess(false);
    
    setTimeout(() => {
      setIsSaving(false);
      setSaveSuccess(true);
      
      // Reset success message after 3 seconds
      setTimeout(() => {
        setSaveSuccess(false);
      }, 3000);
    }, 1000);
  };

  const handleGoBack = (): void => {
    window.location.href = "/dashboard";
  };

  if (!isAuthenticated || loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <Header />
      <main className="container mx-auto py-8 px-4 flex-1">
        <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 mb-8">
          <div>
            <h1 className="text-2xl font-bold">Edit Deployment {id}</h1>
            <p className="text-muted-foreground">Update your code and package files</p>
          </div>
          <div className="flex items-center gap-3">
            <button
              onClick={handleGoBack}
              className="btn btn-outline"
            >
              Cancel
            </button>
            <button
              onClick={handleSave}
              disabled={isSaving}
              className="btn btn-primary relative"
            >
              {isSaving ? (
                <>
                  <span className="animate-spin h-4 w-4 mr-2 border-b-2 border-current rounded-full"></span>
                  Saving...
                </>
              ) : saveSuccess ? (
                <>
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" viewBox="0 0 24 24" fill="none" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                  Saved
                </>
              ) : (
                'Save Changes'
              )}
            </button>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          <div className="space-y-4">
            <div className="flex justify-between items-center">
              <h2 className="text-lg font-semibold">Package.json</h2>
              <span className="text-xs text-muted-foreground">Dependencies configuration</span>
            </div>
            <CodeEditor
              language="json"
              value={packageFile}
              onChange={setPackageFile}
            />
          </div>

          <div className="space-y-4">
            <div className="flex justify-between items-center">
              <h2 className="text-lg font-semibold">Code File</h2>
              <span className="text-xs text-muted-foreground">JavaScript function code</span>
            </div>
            <CodeEditor
              language="javascript"
              value={codeFile}
              onChange={setCodeFile}
            />
          </div>
        </div>
      </main>
    </div>
  );
};

export default EditDeployment;