export interface Deployment {
  id: string;
  name: string;
  status: 'Running' | 'Stopped' | 'Failed' | 'Building';
  createdAt: string;
  language: 'nodejs' | 'go' | 'python';
}

export interface DeploymentFiles {
  packageFile: string;
  codeFile: string;
}

export interface DeploymentFileMap {
  [id: string]: DeploymentFiles;
}