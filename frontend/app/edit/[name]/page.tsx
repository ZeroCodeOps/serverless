"use client";

import { useState, useEffect } from "react";
import Header from "@/components/Header";
import CodeEditor from "@/components/Editor";
import { useAuth } from "@/utils/auth";
import { mockFiles } from "@/utils/mockData";
import { NextPage } from "next";
import { BACKEND_URL } from "@/lib/utils";

const getPackageFileName = (language: string): string => {
  switch (language) {
    case "node":
      return "package.json";
    case "go":
      return "go.mod";
    case "python":
      return "requirements.txt";
    default:
      return "package.json";
  }
};

const EditDeployment: NextPage = () => {
  const [name, setName] = useState<string>("");
  const [deployment, setDeployment] = useState<any>(null);
  const [packageFile, setPackageFile] = useState<string>("");
  const [codeFile, setCodeFile] = useState<string>("");
  const [loading, setLoading] = useState<boolean>(true);
  const [isSaving, setIsSaving] = useState<boolean>(false);
  const [saveSuccess, setSaveSuccess] = useState<boolean>(false);
  const isAuthenticated = useAuth();

  console.log(deployment);

  useEffect(() => {
    setName(window.location.pathname.split("/").pop() || "");
  }, []);

  useEffect(() => {
    if (isAuthenticated && name && typeof name === "string") {
      const fetchData = async () => {
        const response = await fetch(`${BACKEND_URL}/deployments/${name}`);
        const data = await response.json();
        if (response.ok) {
          setDeployment(data);
          setCodeFile(data.code);
          setPackageFile(data.package);
        }
      };
      // Simulate API call
      setTimeout(() => {
        fetchData();
        setLoading(false);
      }, 600);
    }
  }, [isAuthenticated, name]);

  const handleSave = (): void => {
    // Simulate API call
    setIsSaving(true);
    setSaveSuccess(false);
    handleUpload(codeFile, packageFile);

    setTimeout(() => {
      setIsSaving(false);
      setSaveSuccess(true);

      // Reset success message after 3 seconds
      setTimeout(() => {
        setSaveSuccess(false);
      }, 3000);
    }, 1000);
  };

  const handleUpload = async (
    codeFile: string,
    packageFile: string,
  ): Promise<void> => {
    const formData = new FormData();
    const codeBlob = new Blob([codeFile], { type: "text/plain" });
    const packageBlob = new Blob([packageFile], { type: "text/plain" });
    formData.append("code", codeBlob, "main.go");
    formData.append("package", packageBlob, "go.mod");
    const response = await fetch(`${BACKEND_URL}/upload/${name}`, {
      method: "POST",
      body: formData,
    });
    if (response.ok) {
      alert("File uploaded successfully");
    } else {
      alert("Failed to upload file");
    }
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
            <h1 className="text-2xl font-bold">Edit {name}</h1>
            <p className="text-muted-foreground">
              Update your code and package files
            </p>
          </div>
          <div className="flex items-center gap-3">
            <button onClick={handleGoBack} className="btn btn-outline">
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
                      d="M5 13l4 4L19 7"
                    />
                  </svg>
                  Saved
                </>
              ) : (
                "Save Changes"
              )}
            </button>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          <div className="space-y-4">
            <div className="flex justify-between items-center">
              <h2 className="text-lg font-semibold">
                {getPackageFileName(deployment?.language)}
              </h2>
              <span className="text-xs text-muted-foreground">
                Dependencies configuration
              </span>
            </div>
            <CodeEditor
              language={getPackageFileName(deployment?.language).split(".")[1]}
              value={packageFile}
              onChange={setPackageFile}
            />
          </div>

          <div className="space-y-4">
            <div className="flex justify-between items-center">
              <h2 className="text-lg font-semibold">Code File</h2>
              <span className="text-xs text-muted-foreground">
                {deployment?.language} function code
              </span>
            </div>
            <CodeEditor
              language={deployment?.language === "node" ? "javascript" : deployment?.language}
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
